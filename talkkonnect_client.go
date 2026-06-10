package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

type uiChannelUser struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Self   bool   `json:"self"`
}

type uiChannelNode struct {
	Name      string `json:"name"`
	Depth     int    `json:"depth"`
	UserCount int    `json:"userCount"`
	Active    bool   `json:"active"`
}

type uiLastMessage struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
}

type talkkonnectStatus struct {
	Connected     bool            `json:"connected"`
	Transmitting  bool            `json:"transmitting"`
	ServerName    string          `json:"serverName"`
	Server        string          `json:"server"`
	Channel       string          `json:"channel"`
	UsersOnline   int             `json:"usersOnline"`
	ChannelUsers  []uiChannelUser `json:"channelUsers"`
	ChannelTree   []uiChannelNode `json:"channelTree"`
	Receiving     bool            `json:"receiving"`
	LastSpeaker   string          `json:"lastSpeaker"`
	LastMessage   uiLastMessage   `json:"lastMessage"`
	RXVolume      int             `json:"rxVolume"`
	Muted         bool            `json:"muted"`
	InternetRadio struct {
		Enabled      bool   `json:"enabled"`
		Playing      bool   `json:"playing"`
		Status       string `json:"status"`
		StationName  string `json:"stationName"`
		StationIndex int    `json:"stationIndex"`
		StationCount int    `json:"stationCount"`
		Volume       int    `json:"volume"`
	} `json:"internetRadio"`
	IPAddress      string `json:"ipAddress"`
	Bitrate        string `json:"bitrate"`
	UptimeSec      int64  `json:"uptimeSec"`
	Activity       string `json:"activity"`
	MumbleUsername string `json:"mumbleUsername"`
	Version        string `json:"version"`
}

type talkkonnectClient struct {
	url    string
	client *http.Client
}

var channelTreeNodePool = sync.Pool{
	New: func() any {
		s := make([]ChannelTreeNode, 0, 32)
		return &s
	},
}

func acquireChannelTreeNodes(minCap int) []ChannelTreeNode {
	p := channelTreeNodePool.Get().(*[]ChannelTreeNode)
	s := *p
	if cap(s) < minCap {
		channelTreeNodePool.Put(p)
		return make([]ChannelTreeNode, 0, minCap)
	}
	return s[:0]
}

// releaseChannelTreeNodes returns a pooled slice to channelTreeNodePool after zeroing elements.
func releaseChannelTreeNodes(s []ChannelTreeNode) {
	if s == nil {
		return
	}
	for i := range s {
		s[i] = ChannelTreeNode{}
	}
	s = s[:0]
	channelTreeNodePool.Put(&s)
}

func newTalkkonnectClient(url string) *talkkonnectClient {
	return &talkkonnectClient{
		url: strings.TrimSpace(url),
		// No Client.Timeout: deadlines and cancellation flow from context.Context.
		client: &http.Client{},
	}
}

func (c *talkkonnectClient) fetch(ctx context.Context) (talkkonnectStatus, error) {
	var st talkkonnectStatus
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return st, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return st, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return st, fmt.Errorf("talkkonnect status HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return st, err
	}
	return st, nil
}

func (st talkkonnectStatus) toDisplayState() DisplayState {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "talkkonnect"
	}

	out := DisplayState{
		DeviceName:         strings.ToUpper(hostname),
		DeviceIP:           st.IPAddress,
		ServerName:         st.ServerName,
		ServerIP:           st.Server,
		Channel:            strings.ToUpper(st.Channel),
		UserCount:          st.UsersOnline,
		LastSpeaker:        st.LastSpeaker,
		Volume:             st.RXVolume,
		Activity:           st.Activity,
		Receiving:          st.Receiving,
		Connected:          st.Connected,
		Transmitting:       st.Transmitting,
		Muted:              st.Muted,
		MumbleUsername:     strings.TrimSpace(st.MumbleUsername),
		LastMessageSender:  strings.TrimSpace(st.LastMessage.Sender),
		LastMessageText:    strings.TrimSpace(st.LastMessage.Text),
		Mode:               "normal",
		RTT:                "--",
		TalkkonnectVersion: strings.TrimSpace(st.Version),
	}

	switch {
	case st.Transmitting:
		out.TXRXStatus = "T R A N S M I T T I N G "
		out.Activity = "tx"
	case st.Receiving:
		out.TXRXStatus = "R E C E I V I N G "
		out.Activity = "rx"
	case st.Connected:
		out.TXRXStatus = "Idle"
	default:
		out.TXRXStatus = "Offline"
	}

	if st.InternetRadio.Playing || (st.InternetRadio.Enabled && st.InternetRadio.Status == "ducking") {
		out.Channel = strings.ToUpper(st.InternetRadio.StationName)
		if out.Channel == "" {
			out.Channel = "INTERNET RADIO"
		}
		out.ServerName = "Internet Radio"
		out.Volume = st.InternetRadio.Volume
		out.Activity = "radio"
		if st.InternetRadio.Status == "playing" {
			out.TXRXStatus = "STREAM"
		}
	}

	out.ChannelTree = mapChannelTree(st.ChannelTree)
	out.Users = mapChannelUsers(st.ChannelUsers)
	if len(out.Users) > 0 {
		out.UserCount = len(out.Users)
	} else {
		out.UserCount = st.UsersOnline
	}
	return out
}

func mapChannelTree(from []uiChannelNode) []ChannelTreeNode {
	if len(from) == 0 {
		return nil
	}
	out := acquireChannelTreeNodes(len(from))
	for _, n := range from {
		name := strings.TrimSpace(n.Name)
		if name == "" {
			continue
		}
		out = append(out, ChannelTreeNode{
			Name:      name,
			Depth:     n.Depth,
			UserCount: n.UserCount,
			Active:    n.Active,
		})
	}
	return out
}

func mapChannelUsers(from []uiChannelUser) []ChannelUser {
	if len(from) == 0 {
		return nil
	}
	out := make([]ChannelUser, 0, len(from))
	for _, u := range from {
		name := strings.TrimSpace(u.Name)
		if name == "" {
			continue
		}
		status := u.Status
		if status == "" {
			status = "idle"
		}
		out = append(out, ChannelUser{Name: name, Status: status})
	}
	return out
}
