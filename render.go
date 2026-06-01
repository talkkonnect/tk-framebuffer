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
	DeviceName         string
	DeviceIP           string
	ServerName         string
	ServerIP           string
	Channel            string
	ChannelTree        []ChannelTreeNode
	Users              []ChannelUser
	UserCount          int
	TXRXStatus         string
	Mode               string
	LastSpeaker        string
	Elapsed            string
	ActivityEndTime    string // wall-clock time when last TX or RX session ended
	Volume             int
	Muted              bool
	RTT                string
	Activity           string
	Receiving          bool
	Connected          bool
	Transmitting       bool
	Offline            bool
	MumbleUsername     string
	TalkkonnectVersion string
}

const graphicsVersion = "1.01"

var (
	colBlack       = color.RGBA{0, 0, 0, 255}
	colBackground  = color.RGBA{14, 14, 16, 255}
	colPanel       = color.RGBA{24, 26, 30, 255}
	colPanelHead   = color.RGBA{36, 42, 52, 255}
	colPanelEdge   = color.RGBA{58, 72, 92, 255}
	colBlue        = color.RGBA{72, 132, 196, 255}
	colBlueDim     = color.RGBA{44, 68, 98, 255}
	colGreyText    = color.RGBA{170, 174, 182, 255}
	colWhite       = color.RGBA{236, 238, 242, 255}
	colOrange      = color.RGBA{232, 118, 38, 255}
	colRed         = color.RGBA{210, 55, 65, 255}
	colGreen       = color.RGBA{62, 190, 98, 255}
	colVUDim       = color.RGBA{20, 22, 26, 255}
	colVUGreen     = color.RGBA{50, 170, 70, 255}
	colVUYellow    = color.RGBA{210, 190, 60, 255}
	colLightYellow = color.RGBA{200, 150, 60, 255}
	colVURed       = color.RGBA{200, 55, 50, 255}
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

func isSpeakingStatus(status string) bool {
	return strings.EqualFold(status, "speaking")
}

func isTransmittingSelf(u ChannelUser, transmitting bool, selfName string) bool {
	if !transmitting {
		return false
	}
	selfName = strings.TrimSpace(selfName)
	if selfName == "" {
		return false
	}
	return strings.EqualFold(u.Name, selfName)
}

// promoteTransmittingUser moves the logged-in user to the top while transmitting.
func promoteTransmittingUser(users []ChannelUser, selfName string) []ChannelUser {
	selfName = strings.TrimSpace(selfName)
	if selfName == "" {
		return users
	}
	idx := -1
	for i, u := range users {
		if strings.EqualFold(u.Name, selfName) {
			idx = i
			break
		}
	}
	if idx <= 0 {
		if idx == 0 {
			return users
		}
		return append([]ChannelUser{{Name: selfName, Status: "idle"}}, users...)
	}
	out := make([]ChannelUser, 0, len(users))
	out = append(out, users[idx])
	out = append(out, users[:idx]...)
	out = append(out, users[idx+1:]...)
	return out
}

func userStatusColor(status string) color.Color {
	switch {
	case isSpeakingStatus(status):
		return colGreen
	case strings.EqualFold(status, "whisper"):
		return colGreyText
	case strings.EqualFold(status, "muted"):
		return colRed
	default:
		return colGreyText
	}
}

func drawUserIcon(img draw.Image, x, y int, status string, transmittingSelf bool) {
	switch {
	case transmittingSelf:
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), colGreen)
	case isSpeakingStatus(status):
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), colGreen)
	case strings.EqualFold(status, "muted"):
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), colRed)
	default:
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), colPanel)
	}
}

func vuSegmentColor(position float64) color.Color {
	if position > 0.72 {
		return colVURed
	}
	if position > 0.5 {
		return colVUYellow
	}
	return colVUGreen
}

func drawSegmentedHorizontalTrack(img draw.Image, x, trackY, w, trackH int, borderCol color.Color, level float64) {
	track := image.Rect(x, trackY, x+w, trackY+trackH)
	fillRect(img, track, colVUDim)
	strokeRect(img, track, borderCol, 1)
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	segments := 20
	gap := 2
	innerW := w - 4
	segW := (innerW - gap*(segments-1)) / segments
	if segW < 2 {
		segW = 2
	}
	lit := int(float64(segments) * level)
	if lit > segments {
		lit = segments
	}
	for s := 0; s < segments; s++ {
		sx := x + 2 + s*(segW+gap)
		segRect := image.Rect(sx, trackY+2, sx+segW, trackY+trackH-2)
		if s < lit {
			fillRect(img, segRect, vuSegmentColor(float64(s+1)/float64(segments)))
		}
	}
}

func drawSegmentedHorizontalBar(img draw.Image, x, y, w, trackH int, label string, borderCol color.Color, level float64) int {
	drawText(img, x, y+10, label, colGreyText, sizeSmall)
	trackY := y + 14
	drawSegmentedHorizontalTrack(img, x, trackY, w, trackH, borderCol, level)
	return trackY + trackH + 6
}

func drawSignalBar(img draw.Image, x, y, w, trackH int, level float64) int {
	return drawSegmentedHorizontalBar(img, x, y, w, trackH, "Signal", colPanelEdge, level)
}

func drawVolumeBar(img draw.Image, x, y, w, trackH int, volume int, muted bool) int {
	labelCol := colWhite
	if muted {
		labelCol = colRed
	}
	drawText(img, x, y+10, "Speaker Volume", labelCol, sizeSmall)
	drawTextRight(img, x+w, y+10, fmt.Sprintf("%d", volume), colWhite, sizeSmall)
	trackY := y + 14
	drawSegmentedHorizontalTrack(img, x, trackY, w, trackH, colBlueDim, float64(volume)/100.0)
	return trackY + trackH + 6
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
	if st.Transmitting {
		users = promoteTransmittingUser(users, st.MumbleUsername)
	}
	y := listTop
	shown := 0
	for _, u := range users {
		txSelf := isTransmittingSelf(u, st.Transmitting, st.MumbleUsername)
		speaking := isSpeakingStatus(u.Status)
		rowH := 22
		textSize := sizeSmall
		if speaking || txSelf {
			rowH = 28
			textSize = sizeBody
		}
		if y+rowH > listMaxY {
			break
		}
		drawUserIcon(img, col2.Min.X+10, y, u.Status, txSelf)
		line := u.Name
		lineCol := userStatusColor(u.Status)
		if txSelf {
			lineCol = colGreen
		}
		drawText(img, col2.Min.X+28, y+rowH-8, line, lineCol, textSize)
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
	txCol := colBlack
	fillColor := colLightYellow

	if txrx == "" {
		txrx = "O N L I N E"
		fillColor = colLightYellow
	}
	if txrx == "Idle" {
		txrx = "O N L I N E"
		fillColor = colLightYellow
	}
	if st.Transmitting {
		txCol = colBlack
		fillColor = colVURed
	}
	if st.Receiving {
		txCol = colBlack
		fillColor = colVUGreen
	}

	statusBox := image.Rect(col3.Min.X+8, col3.Min.Y+34, col3.Max.X-8, col3.Min.Y+58)
	fillRect(img, statusBox, colVUDim)
	strokeRect(img, statusBox, colPanelEdge, 2)
	fillRect(img, statusBox, fillColor)
	drawText(img, col3.Min.X+14, col3.Min.Y+52, ""+txrx, txCol, sizeBody)
	drawTalkkonnectStatusLED(img, statusBox, talkkonnectOK, now)
	modeY := col3.Min.Y + 66
	drawOutlinedButton(img, image.Rect(col3.Min.X+8, modeY, col3.Min.X+col3.Dx()/2-2, modeY+28), "Broadcast", st.Mode != "whisper")
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

	barsX := col3.Min.X + 8
	barsW := col3.Dx() - 16
	barTrackH := 12

	barsY := modeY + 250

	barsY = drawSignalBar(img, barsX, barsY, barsW, barTrackH, signal)
	volume := st.Volume
	if st.Muted {
		volume = 0
	}
	drawVolumeBar(img, barsX, barsY, barsW, barTrackH, volume, st.Muted)

	// --- Footer ---
	fillRect(img, image.Rect(0, height-footerH, width, height), colPanel)
	strokeRect(img, image.Rect(0, height-footerH, width, height), colPanelEdge, 1)
	footerVersion := fmt.Sprintf("talKKonnect %s  Graphics %s", talkkonnectVersionLabel(st.TalkkonnectVersion), graphicsVersion)
	footerText := "talkkonnect by Suvir Kumar (Released under the MPL License)"
	drawText(img, 5, height-8, footerVersion, colGreyText, sizeLabel)
	drawText(img, 240, height-8, footerText, colGreyText, sizeLabel)
	drawText(img, width-margin-160, height-8, now.Format(time.ANSIC), colGreyText, sizeLabel)
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
		TXRXStatus:         "StandBy",
		Mode:               "Broadcast",
		LastSpeaker:        "suvir",
		Elapsed:            "08s",
		ActivityEndTime:    "14:32:05",
		Volume:             72,
		Muted:              false,
		RTT:                "18ms",
		Activity:           "idle",
		Connected:          true,
		MumbleUsername:     "talkkonnect-demo",
		TalkkonnectVersion: "4.06.03",
	}
}

func talkkonnectVersionLabel(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "—"
	}
	return version
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
	st.TXRXStatus = "Offline"
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
