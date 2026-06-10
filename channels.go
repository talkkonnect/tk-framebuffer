package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// ChannelTreeNode is one row in the server channel tree (inactive rows keep depth).
type ChannelTreeNode struct {
	Name       string
	Depth      int
	UserCount  int
	Active     bool
	Accessible bool
}

func channelDisplayName(name, emptyPlaceholder string) string {
	if name == "" {
		return emptyPlaceholder
	}
	return name
}

func inactiveChannelLabel(n ChannelTreeNode, emptyPlaceholder string) string {
	return fmt.Sprintf("%s (%d)", channelDisplayName(n.Name, emptyPlaceholder), n.UserCount)
}

// partitionChannelTree pulls the active channel out of the hierarchy for pinned display.
func partitionChannelTree(nodes []ChannelTreeNode) (active *ChannelTreeNode, inactive []ChannelTreeNode) {
	inactive = make([]ChannelTreeNode, 0, len(nodes))
	for i := range nodes {
		if nodes[i].Active {
			if active == nil {
				n := nodes[i]
				active = &n
			}
			continue
		}
		inactive = append(inactive, nodes[i])
	}
	return active, inactive
}

func channelInactiveColor(n ChannelTreeNode, pal ThemePalette) color.Color {
	if n.Accessible {
		return pal.GreyText
	}
	return pal.Orange
}

func drawChannelTree(img draw.Image, panel image.Rectangle, nodes []ChannelTreeNode, cfg *UIConfig) {
	x := panel.Min.X + 8
	y := panel.Min.Y + cfg.Layout.PanelContentTop
	maxY := panel.Max.Y - 8
	empty := cfg.Captions.EmptyPlaceholder
	pal := cfg.Palette
	caps := cfg.Captions
	const (
		activeRowH   = 14
		inactiveRowH = 16
	)

	active, rest := partitionChannelTree(nodes)

	if active != nil {
		if y > maxY {
			return
		}
		drawText(img, x, y+13, channelDisplayName(active.Name, empty), pal.Green, sizeChannelActive)
		y += activeRowH
	}

	for _, n := range rest {
		if y > maxY {
			drawText(img, panel.Max.X-24, maxY, caps.ScrollIndicator, pal.GreyText, sizeSmall)
			return
		}
		indent := x + n.Depth*14
		drawText(img, indent, y+14, inactiveChannelLabel(n, empty), channelInactiveColor(n, pal), sizeSmall)
		y += inactiveRowH
	}
}
