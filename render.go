package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"sync"
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
	LastMessageSender  string
	LastMessageText    string
	TalkkonnectVersion string
}

var channelUserPool = sync.Pool{
	New: func() any {
		s := make([]ChannelUser, 0, 32)
		return &s
	},
}

func acquireChannelUsers(minCap int) []ChannelUser {
	p := channelUserPool.Get().(*[]ChannelUser)
	s := *p
	if cap(s) < minCap {
		channelUserPool.Put(p)
		return make([]ChannelUser, 0, minCap)
	}
	return s[:0]
}

// releaseChannelUsers returns a pooled slice to channelUserPool after zeroing elements.
func releaseChannelUsers(s []ChannelUser) {
	if s == nil {
		return
	}
	for i := range s {
		s[i] = ChannelUser{}
	}
	s = s[:0]
	channelUserPool.Put(&s)
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

func drawPanel(img draw.Image, r image.Rectangle, cfg *UIConfig) {
	fillRect(img, r, cfg.Palette.Panel)
	strokeRect(img, r, cfg.Palette.PanelEdge, 1)
}

func drawColumnHeader(img draw.Image, r image.Rectangle, title string, cfg *UIConfig) {
	fillRect(img, r, cfg.Palette.PanelHead)
	strokeRect(img, r, cfg.Palette.Blue, 1)
	drawText(img, r.Min.X+8, r.Min.Y+18, title, cfg.Palette.Blue, sizeLabel)
}

func drawOutlinedButton(img draw.Image, r image.Rectangle, label string, active bool, cfg *UIConfig) {
	c := cfg.Palette.PanelEdge
	if active {
		c = cfg.Palette.Green
	}
	strokeRect(img, r, c, 2)
	drawText(img, r.Min.X+10, r.Min.Y+(r.Dy()+12)/2, label, cfg.Palette.GreyText, sizeLabel)
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
// When pooled is true, the caller must call releaseChannelUsers on out after use.
func promoteTransmittingUser(users []ChannelUser, selfName string) (out []ChannelUser, pooled bool) {
	selfName = strings.TrimSpace(selfName)
	if selfName == "" {
		return users, false
	}
	idx := -1
	for i, u := range users {
		if strings.EqualFold(u.Name, selfName) {
			idx = i
			break
		}
	}
	if idx == 0 {
		return users, false
	}
	needCap := len(users)
	if idx < 0 {
		needCap = len(users) + 1
	}
	out = acquireChannelUsers(needCap)
	pooled = true
	if idx < 0 {
		out = append(out, ChannelUser{Name: selfName, Status: "idle"})
		out = append(out, users...)
		return out, pooled
	}
	out = append(out, users[idx])
	out = append(out, users[:idx]...)
	out = append(out, users[idx+1:]...)
	return out, pooled
}

func userStatusColor(status string, cfg *UIConfig) color.Color {
	switch {
	case isSpeakingStatus(status):
		return cfg.Palette.Green
	case strings.EqualFold(status, "whisper"):
		return cfg.Palette.GreyText
	case strings.EqualFold(status, "muted"):
		return cfg.Palette.Red
	default:
		return cfg.Palette.GreyText
	}
}

func drawUserIcon(img draw.Image, x, y int, status string, transmittingSelf bool, cfg *UIConfig) {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		drawUserIconFallback(img, x, y, status, transmittingSelf, cfg)
		return
	}
	var tile []byte
	switch {
	case transmittingSelf, isSpeakingStatus(status):
		tile = tileUserIconGreen
	case strings.EqualFold(status, "muted"):
		tile = tileUserIconRed
	default:
		tile = tileUserIconPanel
	}
	blitRGBATile(rgba, x+2, y+5, userIconW, userIconH, tile)
}

func drawUserIconFallback(img draw.Image, x, y int, status string, transmittingSelf bool, cfg *UIConfig) {
	switch {
	case transmittingSelf:
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), cfg.Palette.Green)
	case isSpeakingStatus(status):
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), cfg.Palette.Green)
	case strings.EqualFold(status, "muted"):
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), cfg.Palette.Red)
	default:
		fillRect(img, image.Rect(x+2, y+5, x+9, y+15), cfg.Palette.Panel)
	}
}

func vuSegmentColor(position float64, cfg *UIConfig) color.Color {
	if position > 0.72 {
		return cfg.Palette.VURed
	}
	if position > 0.5 {
		return cfg.Palette.VUYellow
	}
	return cfg.Palette.VUGreen
}

func drawSegmentedHorizontalTrack(img draw.Image, x, trackY, w, trackH int, borderCol color.Color, level float64, cfg *UIConfig) {
	track := image.Rect(x, trackY, x+w, trackY+trackH)
	fillRect(img, track, cfg.Palette.VUDim)
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
			fillRect(img, segRect, vuSegmentColor(float64(s+1)/float64(segments), cfg))
		}
	}
}

func drawSegmentedHorizontalBar(img draw.Image, x, y, w, trackH int, label string, borderCol color.Color, level float64, cfg *UIConfig) int {
	drawText(img, x, y+10, label, cfg.Palette.GreyText, sizeSmall)
	trackY := y + 14
	drawSegmentedHorizontalTrack(img, x, trackY, w, trackH, borderCol, level, cfg)
	return trackY + trackH + 6
}

func drawVolumeBar(img draw.Image, x, y, w, trackH int, volume int, muted bool, cfg *UIConfig) int {
	caps := cfg.Captions
	if !muted {
		drawText(img, x, y+10, caps.SpeakerVolumeLabel, cfg.Palette.White, sizeSmall)
		drawTextRight(img, x+w, y+10, fmt.Sprintf("%d", volume), cfg.Palette.White, sizeSmall)
	} else {
		drawText(img, x, y+10, caps.SpeakerMutedLabel, cfg.Palette.Red, sizeSmall)
	}
	trackY := y + 14
	drawSegmentedHorizontalTrack(img, x, trackY, w, trackH, cfg.Palette.BlueDim, float64(volume)/100.0, cfg)
	return trackY + trackH + 6
}

func renderFrame(img draw.Image, width, height int, st DisplayState, signalBars int, talkkonnectOK bool, now time.Time, cfg *UIConfig) {
	lay := cfg.Layout
	caps := cfg.Captions
	pal := cfg.Palette

	fillRect(img, img.Bounds(), pal.Background)

	headerH := lay.HeaderHeight
	footerH := lay.FooterHeight
	margin := lay.Margin
	gap := lay.Gap

	// --- Header ---
	fillRect(img, image.Rect(0, 0, width, headerH), pal.Panel)
	strokeRect(img, image.Rect(0, 0, width, headerH), pal.PanelEdge, 1)

	drawText(img, margin, 18, caps.HostNameLabel+" "+st.DeviceName, pal.GreyText, sizeLabel)
	drawText(img, margin, 36, caps.HostIPLabel+" "+st.DeviceIP, pal.GreyText, sizeLabel)

	srvTitle := st.ServerName
	if srvTitle == "" {
		srvTitle = caps.NotConnected
	}
	mumbleUser := strings.TrimSpace(st.MumbleUsername)
	if mumbleUser == "" {
		mumbleUser = caps.EmptyPlaceholder
	}

	drawText(img, width/2-lay.ServerInfoCenterOffset, 18, caps.ServerNameLabel+" "+srvTitle, pal.GreyText, sizeLabel)
	drawText(img, width/2-lay.ServerInfoCenterOffset, 32, caps.ServerIPLabel+" "+st.ServerIP, pal.GreyText, sizeLabel)
	userLine := caps.UserPrefix + " " + mumbleUser
	drawTextRight(img, width/2+lay.UserLineCenterOffset, 48, userLine, pal.GreyText, sizeLabel)
	//	drawTextRight(img, width-margin, 18, userLine, colWhite, sizeLabel)

	// --- Main 3 columns ---
	bodyTop := headerH + margin
	bodyBottom := height - footerH - margin
	colW := (width - margin*2 - gap*2) / 3
	col1 := image.Rect(margin, bodyTop, margin+colW, bodyBottom)
	col2 := image.Rect(margin+colW+gap, bodyTop, margin+colW*2+gap, bodyBottom)
	col3 := image.Rect(margin+colW*2+gap*2, bodyTop, width-margin, bodyBottom)

	// Left: CHANNELS (tree)
	drawPanel(img, col1, cfg)
	drawColumnHeader(img, image.Rect(col1.Min.X, col1.Min.Y, col1.Max.X, col1.Min.Y+lay.ColumnHeaderHeight), caps.ChannelListTitle, cfg)
	tree := st.ChannelTree
	if len(tree) == 0 && st.Channel != "" {
		tree = []ChannelTreeNode{{
			Name:      st.Channel,
			Depth:     0,
			UserCount: st.UserCount,
			Active:    true,
		}}
	}
	drawChannelTree(img, col1, tree, cfg)

	// Middle: USERS IN CHANNEL
	drawPanel(img, col2, cfg)
	userTitle := fmt.Sprintf(caps.UsersInChannelTitle, st.UserCount)
	drawColumnHeader(img, image.Rect(col2.Min.X, col2.Min.Y, col2.Max.X, col2.Min.Y+lay.ColumnHeaderHeight), userTitle, cfg)
	listTop := col2.Min.Y + lay.PanelContentTop
	listMaxY := col2.Max.Y - 10
	users := st.Users
	if st.Transmitting {
		var promotedPooled bool
		users, promotedPooled = promoteTransmittingUser(users, st.MumbleUsername)
		if promotedPooled {
			defer releaseChannelUsers(users)
		}
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
		drawUserIcon(img, col2.Min.X+10, y, u.Status, txSelf, cfg)
		line := u.Name
		lineCol := userStatusColor(u.Status, cfg)
		if txSelf {
			lineCol = pal.Green
		}
		drawText(img, col2.Min.X+28, y+rowH-8, line, lineCol, textSize)
		y += rowH
		shown++
	}
	if len(st.Users) > shown {
		drawText(img, col2.Max.X-30, col2.Max.Y-10, caps.ScrollIndicator, pal.GreyText, sizeSmall)
	}

	// Right: STATUS & MODE
	drawPanel(img, col3, cfg)
	drawColumnHeader(img, image.Rect(col3.Min.X, col3.Min.Y, col3.Max.X, col3.Min.Y+lay.ColumnHeaderHeight), caps.StatusPanelTitle, cfg)

	txrx := st.TXRXStatus
	txCol := pal.Black
	fillColor := pal.LightYellow

	if txrx == "" {
		txrx = caps.OnlineStatus
		fillColor = pal.LightYellow
	}
	if txrx == "Idle" {
		txrx = caps.OnlineStatus
		fillColor = pal.LightYellow
	}
	if st.Transmitting {
		txCol = pal.Black
		fillColor = pal.VURed
	}
	if st.Receiving {
		txCol = pal.Black
		fillColor = pal.VUGreen
	}

	statusBox := image.Rect(col3.Min.X+8, col3.Min.Y+34, col3.Max.X-8, col3.Min.Y+58)
	fillRect(img, statusBox, pal.VUDim)
	strokeRect(img, statusBox, pal.PanelEdge, 2)
	fillRect(img, statusBox, fillColor)
	drawText(img, col3.Min.X+14, col3.Min.Y+52, txrx, txCol, sizeBody)
	drawTalkkonnectStatusLED(img, statusBox, talkkonnectOK, now, cfg)
	modeY := col3.Min.Y + 66
	drawOutlinedButton(img, image.Rect(col3.Min.X+8, modeY, col3.Min.X+col3.Dx()/2-2, modeY+28), caps.BroadcastLabel, st.Mode != "whisper", cfg)
	drawOutlinedButton(img, image.Rect(col3.Min.X+col3.Dx()/2+2, modeY, col3.Max.X-8, modeY+28), caps.WhisperLabel, st.Mode == "whisper", cfg)
	speaker := st.LastSpeaker
	if speaker == "" {
		speaker = "-"
	}
	drawText(img, col3.Min.X+10, modeY+48, caps.SpeakingLabel+" "+speaker, pal.GreyText, sizeLabel)
	elapsed := st.Elapsed
	if elapsed == "" {
		elapsed = "-"
	}
	drawText(img, col3.Min.X+10, modeY+64, caps.ElapsedLabel+" "+elapsed, pal.GreyText, sizeLabel)
	activityEnd := st.ActivityEndTime
	drawText(img, col3.Min.X+10, modeY+80, caps.LastActivityLabel+" "+activityEnd, pal.GreyText, sizeLabel)

	msgTop := modeY + 98
	msgBottom := modeY + lay.SignalBarsYOffset - 6
	if msgBottom > msgTop+20 {
		drawMumbleMessagePanel(img, image.Rect(col3.Min.X+8, msgTop, col3.Max.X-8, msgBottom), st, cfg)
	}

	barsX := col3.Min.X + 8
	barsW := col3.Dx() - 16
	barTrackH := lay.VolumeBarTrackHeight

	barsY := modeY + lay.SignalBarsYOffset

	barsY = drawSignalMeter(img, barsX, barsY, barsW, signalBars, cfg)
	volume := st.Volume
	if st.Muted {
		volume = 0
	}
	drawVolumeBar(img, barsX, barsY, barsW, barTrackH, volume, st.Muted, cfg)

	// --- Footer ---
	fillRect(img, image.Rect(0, height-footerH, width, height), pal.Panel)
	strokeRect(img, image.Rect(0, height-footerH, width, height), pal.PanelEdge, 1)
	footerVersion := fmt.Sprintf(caps.FooterVersionFormat, talkkonnectVersionLabel(st.TalkkonnectVersion, caps.EmptyPlaceholder), caps.GraphicsVersion)
	drawText(img, lay.FooterVersionX, height-8, footerVersion, pal.GreyText, sizeLabel)
	drawText(img, lay.FooterTextX, height-8, caps.FooterCredit, pal.GreyText, sizeLabel)
	drawText(img, width-margin-lay.FooterClockRightMargin, height-8, now.Format(time.ANSIC), pal.GreyText, sizeLabel)
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
		LastMessageSender:  "Zoran",
		LastMessageText:    "สวัสดีครับ พบกันที่ช่อง General ใน 5 นาที",
		TalkkonnectVersion: "4.06.03",
	}
}

func talkkonnectVersionLabel(version, emptyPlaceholder string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return emptyPlaceholder
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
	st.Elapsed = "-"
	st.ActivityEndTime = ""
	st.Offline = true
	st.Connected = false
	st.MumbleUsername = ""
	st.LastMessageSender = ""
	st.LastMessageText = ""
	st.RTT = ""
	return st
}

func trimHost(server string) string {
	server = strings.TrimSpace(server)
	if i := strings.Index(server, ":"); i > 0 {
		return server[:i]
	}
	return server
}
