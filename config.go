package main

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// UIConfig holds theme colors, layout dimensions, and UI captions loaded from XML.
type UIConfig struct {
	Colors   Colors   `xml:"colors"`
	Layout   Layout   `xml:"layout"`
	Captions Captions `xml:"captions"`
	Palette  ThemePalette
}

// Colors defines hex color strings in the theme XML.
type Colors struct {
	Black          string `xml:"black"`
	Background     string `xml:"background"`
	Panel          string `xml:"panel"`
	PanelHead      string `xml:"panelHead"`
	PanelEdge      string `xml:"panelEdge"`
	Blue           string `xml:"blue"`
	BlueDim        string `xml:"blueDim"`
	GreyText       string `xml:"greyText"`
	White          string `xml:"white"`
	Orange         string `xml:"orange"`
	Red            string `xml:"red"`
	Green          string `xml:"green"`
	VUDim          string `xml:"vuDim"`
	VUGreen        string `xml:"vuGreen"`
	VUYellow       string `xml:"vuYellow"`
	LightYellow    string `xml:"lightYellow"`
	VURed          string `xml:"vuRed"`
	SignalInactive string `xml:"signalInactive"`
}

// ThemePalette holds parsed RGBA colors resolved once at load time.
type ThemePalette struct {
	Black          color.RGBA
	Background     color.RGBA
	Panel          color.RGBA
	PanelHead      color.RGBA
	PanelEdge      color.RGBA
	Blue           color.RGBA
	BlueDim        color.RGBA
	GreyText       color.RGBA
	White          color.RGBA
	Orange         color.RGBA
	Red            color.RGBA
	Green          color.RGBA
	VUDim          color.RGBA
	VUGreen        color.RGBA
	VUYellow       color.RGBA
	LightYellow    color.RGBA
	VURed          color.RGBA
	SignalInactive color.RGBA
}

// Layout defines pixel dimensions and offsets for the UI chrome.
type Layout struct {
	HeaderHeight           int `xml:"headerHeight"`
	FooterHeight           int `xml:"footerHeight"`
	Margin                 int `xml:"margin"`
	Gap                    int `xml:"gap"`
	ColumnHeaderHeight     int `xml:"columnHeaderHeight"`
	PanelContentTop        int `xml:"panelContentTop"`
	FooterVersionX         int `xml:"footerVersionX"`
	FooterTextX            int `xml:"footerTextX"`
	FooterClockRightMargin int `xml:"footerClockRightMargin"`
	ServerInfoCenterOffset int `xml:"serverInfoCenterOffset"`
	UserLineCenterOffset   int `xml:"userLineCenterOffset"`
	SignalBarsYOffset      int `xml:"signalBarsYOffset"`
	VolumeBarTrackHeight   int `xml:"volumeBarTrackHeight"`
}

// Captions defines static UI label text.
type Captions struct {
	HostNameLabel       string `xml:"hostNameLabel"`
	HostIPLabel         string `xml:"hostIPLabel"`
	ServerNameLabel     string `xml:"serverNameLabel"`
	ServerIPLabel       string `xml:"serverIPLabel"`
	UserPrefix          string `xml:"userPrefix"`
	NotConnected        string `xml:"notConnected"`
	EmptyPlaceholder    string `xml:"emptyPlaceholder"`
	ChannelListTitle    string `xml:"channelListTitle"`
	UsersInChannelTitle string `xml:"usersInChannelTitle"`
	StatusPanelTitle    string `xml:"statusPanelTitle"`
	OnlineStatus        string `xml:"onlineStatus"`
	BroadcastLabel      string `xml:"broadcastLabel"`
	WhisperLabel        string `xml:"whisperLabel"`
	SpeakingLabel       string `xml:"speakingLabel"`
	ElapsedLabel        string `xml:"elapsedLabel"`
	LastActivityLabel   string `xml:"lastActivityLabel"`
	MessageFromFormat   string `xml:"messageFromFormat"`
	MessageFromServerLabel string `xml:"messageFromServerLabel"`
	SpeakerVolumeLabel  string `xml:"speakerVolumeLabel"`
	SpeakerMutedLabel   string `xml:"speakerMutedLabel"`
	SignalLevelLabel    string `xml:"signalLevelLabel"`
	ScrollIndicator     string `xml:"scrollIndicator"`
	FooterVersionFormat string `xml:"footerVersionFormat"`
	FooterCredit        string `xml:"footerCredit"`
	GraphicsVersion     string `xml:"graphicsVersion"`
}

func defaultUIConfig() *UIConfig {
	return &UIConfig{
		Colors: Colors{
			Black:          "#000000",
			Background:     "#0E0E10",
			Panel:          "#181A1E",
			PanelHead:      "#242A34",
			PanelEdge:      "#3A485C",
			Blue:           "#4884C4",
			BlueDim:        "#2C4462",
			GreyText:       "#AAAEB6",
			White:          "#ECEEF2",
			Orange:         "#E87626",
			Red:            "#D23741",
			Green:          "#3EBE62",
			VUDim:          "#14161A",
			VUGreen:        "#32AA46",
			VUYellow:       "#D2BE3C",
			LightYellow:    "#C8963C",
			VURed:          "#C83732",
			SignalInactive: "#2A2E36",
		},
		Layout: Layout{
			HeaderHeight:           54,
			FooterHeight:           34,
			Margin:                 6,
			Gap:                    6,
			ColumnHeaderHeight:     24,
			PanelContentTop:        30,
			FooterVersionX:         5,
			FooterTextX:            240,
			FooterClockRightMargin: 160,
			ServerInfoCenterOffset: 120,
			UserLineCenterOffset:   44,
			SignalBarsYOffset:      250,
			VolumeBarTrackHeight:   12,
		},
		Captions: Captions{
			HostNameLabel:       "HostName:",
			HostIPLabel:         "HostIP:",
			ServerNameLabel:     "Name:",
			ServerIPLabel:       "IP:",
			UserPrefix:          "USER:",
			NotConnected:        "Not Connected",
			EmptyPlaceholder:    "—",
			ChannelListTitle:    "Channel List",
			UsersInChannelTitle: "%d Users In Connected Channel",
			StatusPanelTitle:    "System Status & Communication Mode",
			OnlineStatus:        "O N L I N E",
			BroadcastLabel:      "Broadcast",
			WhisperLabel:        "Whisper",
			SpeakingLabel:       "Speaking:",
			ElapsedLabel:        "Elapsed  :",
			LastActivityLabel:   "Last Activity  :",
			MessageFromFormat:   "From %s:",
			MessageFromServerLabel: "Server",
			SpeakerVolumeLabel:  "Speaker Volume",
			SpeakerMutedLabel:   "Speaker (Muted)",
			SignalLevelLabel:    "Signal Level (RX/TX)",
			ScrollIndicator:     "▼",
			FooterVersionFormat: "talKKonnect %s  Graphics %s",
			FooterCredit:        "talkkonnect by Suvir Kumar (Released under the MPL License)",
			GraphicsVersion:     "1.08",
		},
	}
}

func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q (expected #RRGGBB or RRGGBB)", s)
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("parse red in %q: %w", s, err)
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("parse green in %q: %w", s, err)
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("parse blue in %q: %w", s, err)
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

func (cfg *UIConfig) resolvePalette() error {
	c := cfg.Colors
	var err error
	if cfg.Palette.Black, err = parseHexColor(c.Black); err != nil {
		return err
	}
	if cfg.Palette.Background, err = parseHexColor(c.Background); err != nil {
		return err
	}
	if cfg.Palette.Panel, err = parseHexColor(c.Panel); err != nil {
		return err
	}
	if cfg.Palette.PanelHead, err = parseHexColor(c.PanelHead); err != nil {
		return err
	}
	if cfg.Palette.PanelEdge, err = parseHexColor(c.PanelEdge); err != nil {
		return err
	}
	if cfg.Palette.Blue, err = parseHexColor(c.Blue); err != nil {
		return err
	}
	if cfg.Palette.BlueDim, err = parseHexColor(c.BlueDim); err != nil {
		return err
	}
	if cfg.Palette.GreyText, err = parseHexColor(c.GreyText); err != nil {
		return err
	}
	if cfg.Palette.White, err = parseHexColor(c.White); err != nil {
		return err
	}
	if cfg.Palette.Orange, err = parseHexColor(c.Orange); err != nil {
		return err
	}
	if cfg.Palette.Red, err = parseHexColor(c.Red); err != nil {
		return err
	}
	if cfg.Palette.Green, err = parseHexColor(c.Green); err != nil {
		return err
	}
	if cfg.Palette.VUDim, err = parseHexColor(c.VUDim); err != nil {
		return err
	}
	if cfg.Palette.VUGreen, err = parseHexColor(c.VUGreen); err != nil {
		return err
	}
	if cfg.Palette.VUYellow, err = parseHexColor(c.VUYellow); err != nil {
		return err
	}
	if cfg.Palette.LightYellow, err = parseHexColor(c.LightYellow); err != nil {
		return err
	}
	if cfg.Palette.VURed, err = parseHexColor(c.VURed); err != nil {
		return err
	}
	if cfg.Palette.SignalInactive, err = parseHexColor(c.SignalInactive); err != nil {
		return err
	}
	return nil
}

func mergeUIConfig(base, overlay *UIConfig) {
	if overlay.Colors.Black != "" {
		base.Colors.Black = overlay.Colors.Black
	}
	if overlay.Colors.Background != "" {
		base.Colors.Background = overlay.Colors.Background
	}
	if overlay.Colors.Panel != "" {
		base.Colors.Panel = overlay.Colors.Panel
	}
	if overlay.Colors.PanelHead != "" {
		base.Colors.PanelHead = overlay.Colors.PanelHead
	}
	if overlay.Colors.PanelEdge != "" {
		base.Colors.PanelEdge = overlay.Colors.PanelEdge
	}
	if overlay.Colors.Blue != "" {
		base.Colors.Blue = overlay.Colors.Blue
	}
	if overlay.Colors.BlueDim != "" {
		base.Colors.BlueDim = overlay.Colors.BlueDim
	}
	if overlay.Colors.GreyText != "" {
		base.Colors.GreyText = overlay.Colors.GreyText
	}
	if overlay.Colors.White != "" {
		base.Colors.White = overlay.Colors.White
	}
	if overlay.Colors.Orange != "" {
		base.Colors.Orange = overlay.Colors.Orange
	}
	if overlay.Colors.Red != "" {
		base.Colors.Red = overlay.Colors.Red
	}
	if overlay.Colors.Green != "" {
		base.Colors.Green = overlay.Colors.Green
	}
	if overlay.Colors.VUDim != "" {
		base.Colors.VUDim = overlay.Colors.VUDim
	}
	if overlay.Colors.VUGreen != "" {
		base.Colors.VUGreen = overlay.Colors.VUGreen
	}
	if overlay.Colors.VUYellow != "" {
		base.Colors.VUYellow = overlay.Colors.VUYellow
	}
	if overlay.Colors.LightYellow != "" {
		base.Colors.LightYellow = overlay.Colors.LightYellow
	}
	if overlay.Colors.VURed != "" {
		base.Colors.VURed = overlay.Colors.VURed
	}
	if overlay.Colors.SignalInactive != "" {
		base.Colors.SignalInactive = overlay.Colors.SignalInactive
	}

	if overlay.Layout.HeaderHeight > 0 {
		base.Layout.HeaderHeight = overlay.Layout.HeaderHeight
	}
	if overlay.Layout.FooterHeight > 0 {
		base.Layout.FooterHeight = overlay.Layout.FooterHeight
	}
	if overlay.Layout.Margin > 0 {
		base.Layout.Margin = overlay.Layout.Margin
	}
	if overlay.Layout.Gap > 0 {
		base.Layout.Gap = overlay.Layout.Gap
	}
	if overlay.Layout.ColumnHeaderHeight > 0 {
		base.Layout.ColumnHeaderHeight = overlay.Layout.ColumnHeaderHeight
	}
	if overlay.Layout.PanelContentTop > 0 {
		base.Layout.PanelContentTop = overlay.Layout.PanelContentTop
	}
	if overlay.Layout.FooterVersionX > 0 {
		base.Layout.FooterVersionX = overlay.Layout.FooterVersionX
	}
	if overlay.Layout.FooterTextX > 0 {
		base.Layout.FooterTextX = overlay.Layout.FooterTextX
	}
	if overlay.Layout.FooterClockRightMargin > 0 {
		base.Layout.FooterClockRightMargin = overlay.Layout.FooterClockRightMargin
	}
	if overlay.Layout.ServerInfoCenterOffset > 0 {
		base.Layout.ServerInfoCenterOffset = overlay.Layout.ServerInfoCenterOffset
	}
	if overlay.Layout.UserLineCenterOffset > 0 {
		base.Layout.UserLineCenterOffset = overlay.Layout.UserLineCenterOffset
	}
	if overlay.Layout.SignalBarsYOffset > 0 {
		base.Layout.SignalBarsYOffset = overlay.Layout.SignalBarsYOffset
	}
	if overlay.Layout.VolumeBarTrackHeight > 0 {
		base.Layout.VolumeBarTrackHeight = overlay.Layout.VolumeBarTrackHeight
	}

	mergeCaptions(&base.Captions, &overlay.Captions)
}

func mergeCaptions(base, overlay *Captions) {
	if overlay.HostNameLabel != "" {
		base.HostNameLabel = overlay.HostNameLabel
	}
	if overlay.HostIPLabel != "" {
		base.HostIPLabel = overlay.HostIPLabel
	}
	if overlay.ServerNameLabel != "" {
		base.ServerNameLabel = overlay.ServerNameLabel
	}
	if overlay.ServerIPLabel != "" {
		base.ServerIPLabel = overlay.ServerIPLabel
	}
	if overlay.UserPrefix != "" {
		base.UserPrefix = overlay.UserPrefix
	}
	if overlay.NotConnected != "" {
		base.NotConnected = overlay.NotConnected
	}
	if overlay.EmptyPlaceholder != "" {
		base.EmptyPlaceholder = overlay.EmptyPlaceholder
	}
	if overlay.ChannelListTitle != "" {
		base.ChannelListTitle = overlay.ChannelListTitle
	}
	if overlay.UsersInChannelTitle != "" {
		base.UsersInChannelTitle = overlay.UsersInChannelTitle
	}
	if overlay.StatusPanelTitle != "" {
		base.StatusPanelTitle = overlay.StatusPanelTitle
	}
	if overlay.OnlineStatus != "" {
		base.OnlineStatus = overlay.OnlineStatus
	}
	if overlay.BroadcastLabel != "" {
		base.BroadcastLabel = overlay.BroadcastLabel
	}
	if overlay.WhisperLabel != "" {
		base.WhisperLabel = overlay.WhisperLabel
	}
	if overlay.SpeakingLabel != "" {
		base.SpeakingLabel = overlay.SpeakingLabel
	}
	if overlay.ElapsedLabel != "" {
		base.ElapsedLabel = overlay.ElapsedLabel
	}
	if overlay.LastActivityLabel != "" {
		base.LastActivityLabel = overlay.LastActivityLabel
	}
	if overlay.MessageFromFormat != "" {
		base.MessageFromFormat = overlay.MessageFromFormat
	}
	if overlay.MessageFromServerLabel != "" {
		base.MessageFromServerLabel = overlay.MessageFromServerLabel
	}
	if overlay.SpeakerVolumeLabel != "" {
		base.SpeakerVolumeLabel = overlay.SpeakerVolumeLabel
	}
	if overlay.SpeakerMutedLabel != "" {
		base.SpeakerMutedLabel = overlay.SpeakerMutedLabel
	}
	if overlay.SignalLevelLabel != "" {
		base.SignalLevelLabel = overlay.SignalLevelLabel
	}
	if overlay.ScrollIndicator != "" {
		base.ScrollIndicator = overlay.ScrollIndicator
	}
	if overlay.FooterVersionFormat != "" {
		base.FooterVersionFormat = overlay.FooterVersionFormat
	}
	if overlay.FooterCredit != "" {
		base.FooterCredit = overlay.FooterCredit
	}
	if overlay.GraphicsVersion != "" {
		base.GraphicsVersion = overlay.GraphicsVersion
	}
}

// loadConfig reads theme XML from path. Missing files fall back to built-in defaults.
func loadConfig(path string) (*UIConfig, error) {
	cfg := defaultUIConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Theme file %s not found, using built-in defaults.\n", path)
			if err := cfg.resolvePalette(); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read theme %s: %w", path, err)
	}

	var fileCfg UIConfig
	if err := xml.Unmarshal(data, &fileCfg); err != nil {
		fmt.Printf("Theme file %s parse error: %v — using built-in defaults.\n", path, err)
		if err := cfg.resolvePalette(); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	mergeUIConfig(cfg, &fileCfg)
	if err := cfg.resolvePalette(); err != nil {
		fmt.Printf("Theme file %s color error: %v — using built-in defaults.\n", path, err)
		fallback := defaultUIConfig()
		if err := fallback.resolvePalette(); err != nil {
			return nil, err
		}
		return fallback, nil
	}

	fmt.Printf("Loaded theme from %s\n", path)
	return cfg, nil
}
