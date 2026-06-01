package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
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

func newTalkkonnectClient(url string) *talkkonnectClient {
	return &talkkonnectClient{
		url: strings.TrimSpace(url),
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *talkkonnectClient) fetch() (talkkonnectStatus, error) {
	var st talkkonnectStatus
	resp, err := c.client.Get(c.url)
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
	out := make([]ChannelTreeNode, 0, len(from))
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
