package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"time"
)

type ChannelUser struct {
	Name   string
	Status string // idle, speaking, whisper, mute
}

type DisplayState struct {
	DeviceName   string
	DeviceIP     string
	ServerName   string
	ServerIP     string
	Channel      string
	ChannelTree  []ChannelTreeNode
	Users        []ChannelUser
	UserCount    int
	TXRXStatus   string
	Mode         string
	LastSpeaker  string
	Elapsed         string
	ActivityEndTime string // wall-clock time when last TX or RX session ended
	Volume       int
	Muted        bool
	RTT          string
	Activity     string
	Receiving    bool
	Connected    bool
	Transmitting bool
	Offline        bool
	MumbleUsername string
}

var (
	colBlack = color.RGBA{0, 0, 0, 255}
	colBackground = color.RGBA{14, 14, 16, 255}
	colPanel      = color.RGBA{24, 26, 30, 255}
	colPanelHead  = color.RGBA{36, 42, 52, 255}
	colPanelEdge  = color.RGBA{58, 72, 92, 255}
	colBlue       = color.RGBA{72, 132, 196, 255}
	colBlueDim    = color.RGBA{44, 68, 98, 255}
	colGreyText   = color.RGBA{170, 174, 182, 255}
	colWhite      = color.RGBA{236, 238, 242, 255}
	colOrange     = color.RGBA{232, 118, 38, 255}
	colRed        = color.RGBA{210, 55, 65, 255}
	colGreen      = color.RGBA{62, 190, 98, 255}
	colVUDim      = color.RGBA{20, 22, 26, 255}
	colVUGreen    = color.RGBA{50, 170, 70, 255}
	colVUYellow   = color.RGBA{210, 190, 60, 255}
	colVURed      = color.RGBA{200, 55, 50, 255}
)

func fillRect(img draw.Image, r image.Rectangle, col color.Color) {
	draw.Draw(img, r, &image.Uniform{col}, image.Point{}, draw.Src)
}

func strokeRect(img draw.Image, r image.Rectangle, col color.Color, w int) {
	if w < 1 {
		w = 1
	}
	for i := 0; i < w; i++ {
		fillRect(img, image.Rect(r.Min.X, r.Min.Y+i, r.Max.X, r.Min.Y+i+1), col)
		fillRect(img, image.Rect(r.Min.X, r.Max.Y-1-i, r.Max.X, r.Max.Y-i), col)
		fillRect(img, image.Rect(r.Min.X+i, r.Min.Y, r.Min.X+i+1, r.Max.Y), col)
		fillRect(img, image.Rect(r.Max.X-1-i, r.Min.Y, r.Max.X-i, r.Max.Y), col)
	}
}

func drawPanel(img draw.Image, r image.Rectangle) {
	fillRect(img, r, colPanel)
	strokeRect(img, r, colPanelEdge, 1)
}

func drawColumnHeader(img draw.Image, r image.Rectangle, title string) {
	fillRect(img, r, colPanelHead)
	strokeRect(img, r, colBlue, 1)
	drawText(img, r.Min.X+8, r.Min.Y+18, title, colBlue, sizeLabel)
}

func drawOutlinedButton(img draw.Image, r image.Rectangle, label string, active bool) {
	c := colPanelEdge
	if active {
		c = colOrange
	}
	strokeRect(img, r, c, 2)
	drawText(img, r.Min.X+10, r.Min.Y+(r.Dy()+12)/2, label, colGreyText, sizeLabel)
}

func userStatusColor(status string) color.Color {
	switch status {
	case "Speaking":
		return colGreen
	case "whisper":
		return colGreyText
	case "Muted":
		return colRed
	default:
		return colGreyText
	}
}

func drawUserIcon(img draw.Image, x, y int, status string) {
	switch status {
	case "Speaking":
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), colGreen)
	case "Muted":
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), colRed)
	default:
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), colPanel)
	}
}

func drawSignalBar(img draw.Image, x, y, w, h int, level float64) {
	segments := 20
	gap := 2
	segH := (h - gap*(segments-1)) / segments
	if segH < 2 {
		segH = 2
	}
	drawText(img, x+12, y+8, "Activity Bar", colGreyText, sizeSmall)
	track := image.Rect(x, y+12, x+w, y+12+h)
	fillRect(img, track, colVUDim)
	strokeRect(img, track, colPanelEdge, 1)
	lit := int(float64(segments) * level)
	if lit > segments {
		lit = segments
	}
	for s := 0; s < segments; s++ {
		sy := y + 12 + h - (s+1)*(segH+gap)
		segRect := image.Rect(x+2, sy, x+w-2, sy+segH)
		if s < lit {
			p := float64(s+1) / float64(segments)
			var c color.Color = colVUGreen
			if p > 0.72 {
				c = colVURed
			} else if p > 0.5 {
				c = colVUYellow
			}
			fillRect(img, segRect, c)
		}
	}
}

func drawVerticalVolume(img draw.Image, x, y, w, h int, volume int) {
	drawText(img, x+12, y+8, "RX Volume", colGreyText, sizeSmall)
	track := image.Rect(x, y+12, x+w, y+12+h)
	fillRect(img, track, colVUDim)
	strokeRect(img, track, colBlueDim, 1)
	knobY := y + 12 + h - int(float64(h-16)*float64(volume)/100.0) - 8
	fillRect(img, image.Rect(x+2, knobY, x+w-2, knobY+8), colRed)
	drawText(img, x+25, y+12+h+12, fmt.Sprintf("%d", volume), colWhite, sizeBody)
}
func muteBoxCaption(muted bool) string {
	if muted {
		return "MUTED"
	}
	return "ACTIVE"
}

func drawMuteButton(img draw.Image, r image.Rectangle, muted bool) {
	bg := colGreen
	if muted {
		bg = colRed
	}
	fillRect(img, r, bg)
	strokeRect(img, r, colPanelEdge, 1)
	caption := muteBoxCaption(muted)
	drawTextCentered(img, r, caption, colBlack, sizeBody)
}

func renderFrame(img draw.Image, width, height int, st DisplayState, signal float64, talkkonnectOK bool, now time.Time) {
	fillRect(img, img.Bounds(), colBackground)

	headerH := 54
	footerH := 34
	margin := 6
	gap := 6

	// --- Header ---
	fillRect(img, image.Rect(0, 0, width, headerH), colPanel)
	strokeRect(img, image.Rect(0, 0, width, headerH), colPanelEdge, 1)

	drawText(img, margin, 18, "DEVICE: "+st.DeviceName, colGreyText, sizeLabel)
	drawText(img, margin, 36, "IP: "+st.DeviceIP, colGreyText, sizeLabel)

	srvTitle := st.ServerName
	if srvTitle == "" {
		srvTitle = "Not Connected"
	}
	drawText(img, width/2-120, 18, "SERVER: "+srvTitle, colGreyText, sizeLabel)
	drawText(img, width/2-120, 36, "IP: "+st.ServerIP, colGreyText, sizeLabel)

	mumbleUser := strings.TrimSpace(st.MumbleUsername)
	if mumbleUser == "" {
		mumbleUser = "—"
	}
	userLine := "USER: " + mumbleUser
	drawTextRight(img, width-margin, 18, userLine, colWhite, sizeLabel)
	drawTalkkonnectStatusLED(img, width, margin, talkkonnectOK, now)

	// --- Main 3 columns ---
	bodyTop := headerH + margin
	bodyBottom := height - footerH - margin
	colW := (width - margin*2 - gap*2) / 3
	col1 := image.Rect(margin, bodyTop, margin+colW, bodyBottom)
	col2 := image.Rect(margin+colW+gap, bodyTop, margin+colW*2+gap, bodyBottom)
	col3 := image.Rect(margin+colW*2+gap*2, bodyTop, width-margin, bodyBottom)

	// Left: CHANNELS (tree)
	drawPanel(img, col1)
	drawColumnHeader(img, image.Rect(col1.Min.X, col1.Min.Y, col1.Max.X, col1.Min.Y+24), "Channel List")
	tree := st.ChannelTree
	if len(tree) == 0 && st.Channel != "" {
		tree = []ChannelTreeNode{{
			Name:      st.Channel,
			Depth:     0,
			UserCount: st.UserCount,
			Active:    true,
		}}
	}
	drawChannelTree(img, col1, tree)

	// Middle: USERS IN CHANNEL
	drawPanel(img, col2)
	userTitle := fmt.Sprintf("Users In Current Channel (%d)", st.UserCount)
	drawColumnHeader(img, image.Rect(col2.Min.X, col2.Min.Y, col2.Max.X, col2.Min.Y+24), userTitle)
	listTop := col2.Min.Y + 30
	listMaxY := col2.Max.Y - 10
	users := st.Users
	y := listTop
	shown := 0
	for _, u := range users {
		rowH := 22
		textSize := sizeSmall
		if u.Status == "Speaking" {
			rowH = 28
			textSize = sizeBody
		}
		if y+rowH > listMaxY {
			break
		}
		drawUserIcon(img, col2.Min.X+10, y, u.Status)
		line := u.Name + " [" + u.Status + "]"
		drawText(img, col2.Min.X+28, y+rowH-8, line, userStatusColor(u.Status), textSize)
		y += rowH
		shown++
	}
	if len(st.Users) > shown {
		drawText(img, col2.Max.X-30, col2.Max.Y-10, "▼", colGreyText, sizeSmall)
	}

	// Right: STATUS & MODE
	drawPanel(img, col3)
	drawColumnHeader(img, image.Rect(col3.Min.X, col3.Min.Y, col3.Max.X, col3.Min.Y+24), "Activity Status & Communication Mode")

	txrx := st.TXRXStatus
	if txrx == "" {
		txrx = "STANDBY"
	}
	if txrx == "IDLE" {
		txrx = "STANDBY"
	}
	txCol := colVUYellow
	if st.Transmitting {
		txCol = colRed
	}
	if st.Receiving {
		txCol = colGreen
	}
	statusBox := image.Rect(col3.Min.X+8, col3.Min.Y+30, col3.Max.X-8, col3.Min.Y+58)
	fillRect(img, statusBox, colVUDim)
	strokeRect(img, statusBox, colPanelEdge, 1)
	drawText(img, col3.Min.X+14, col3.Min.Y+52, ""+txrx, txCol, sizeBody)

	modeY := col3.Min.Y + 66
	drawOutlinedButton(img, image.Rect(col3.Min.X+8, modeY, col3.Min.X+col3.Dx()/2-2, modeY+28), "Normal", st.Mode != "whisper")
	drawOutlinedButton(img, image.Rect(col3.Min.X+col3.Dx()/2+2, modeY, col3.Max.X-8, modeY+28), "Whisper", st.Mode == "whisper")
	speaker := st.LastSpeaker
	if speaker == "" {
		speaker = " "
	}
	drawText(img, col3.Min.X+10, modeY+48, "Speaker: "+speaker, colGreyText, sizeLabel)
	elapsed := st.Elapsed
	if elapsed == "" {
		elapsed = "00s"
	}
	drawText(img, col3.Min.X+10, modeY+64, "Elapsed  : "+elapsed, colGreyText, sizeLabel)
	activityEnd := st.ActivityEndTime
	if activityEnd == "" {
		activityEnd = "—"
	}
	drawText(img, col3.Min.X+10, modeY+80, "Activity  : "+activityEnd, colGreyText, sizeLabel)

	audioTop := col3.Min.Y + 170
	audioH := col3.Max.Y - audioTop - 8
	barW := col3.Dx()/3 - 4
	//suvir fix box sizes here
	drawSignalBar(img, col3.Min.X+6, audioTop, barW, audioH-20, signal)
	drawVerticalVolume(img, col3.Min.X+col3.Dx()/3+4, audioTop, barW, audioH-20, st.Volume)
	muteR := image.Rect(col3.Max.X-barW-2, audioTop+12, col3.Max.X-8, audioTop+audioH-8)
	drawMuteButton(img, muteR, st.Muted)

	// --- Footer ---
	fillRect(img, image.Rect(0, height-footerH, width, height), colPanel)
	strokeRect(img, image.Rect(0, height-footerH, width, height), colPanelEdge, 1)
	logoRect := image.Rect(margin, height-footerH+4, margin+140, height-4)
	drawBrandLogo(img, logoRect)
	footerText := "Web: www.talkkonnect.com Facebook: www.facebook.com/talkkonnect Email: suvir@talkkonnect.com"
	drawText(img, 50, height-8, footerText, colVUYellow, sizeLabel)
	drawText(img, width-margin-125, height-8,now.Format(time.ANSIC), colGreyText, sizeSmall)
}

func mockDisplayState() DisplayState {
	return DisplayState{
		DeviceName: "TK-PI4-NODE",
		DeviceIP:   "192.168.1.1",
		ServerName: "mumble.talkkonnect.com",
		ServerIP:   "111.223.36.158",
		Channel:    "HAM-CB",
		ChannelTree: []ChannelTreeNode{
			{Name: "Root", Depth: 0, UserCount: 0, Active: false},
			{Name: "General", Depth: 1, UserCount: 14, Active: true},
			{Name: "Support", Depth: 1, UserCount: 2, Active: false},
			{Name: "Emergency", Depth: 1, UserCount: 0, Active: false},
		},
		UserCount: 14,
		Users: []ChannelUser{
			{Name: "Suvir", Status: "idle"},
			{Name: "Zoran", Status: "Speaking"},
			{Name: "Panajon", Status: "idle"},
			{Name: "mtech", Status: "idle"},
			{Name: "User4", Status: "whisper"},
			{Name: "User5", Status: "idle"},
			{Name: "User6", Status: "idle"},
			{Name: "User7", Status: "idle"},
			{Name: "User8", Status: "idle"},
			{Name: "User9", Status: "Muted"},
		},
		TXRXStatus:  "STANDBY",
		Mode:        "normal",
		LastSpeaker: "suvir",
		Elapsed:         "08s",
		ActivityEndTime: "14:32:05",
		Volume:      72,
		Muted:       false,
		RTT:         "18ms",
		Activity:       "idle",
		Connected:      true,
		MumbleUsername: "talkkonnect-demo",
	}
}

func offlineDisplayState() DisplayState {
	st := mockDisplayState()
	st.DeviceIP = "—"
	st.ServerName = "Offline"
	st.ServerIP = "—"
	st.Channel = "Not Connected"
	st.ChannelTree = nil
	st.Users = nil
	st.UserCount = 0
	st.TXRXStatus = "OFFLINE"
	st.LastSpeaker = "—"
	st.Elapsed = "00s"
	st.ActivityEndTime = ""
	st.Offline = true
	st.Connected = false
	st.MumbleUsername = ""
	st.RTT = "--"
	return st
}

func trimHost(server string) string {
	server = strings.TrimSpace(server)
	if i := strings.Index(server, ":"); i > 0 {
		return server[:i]
	}
	return server
}
