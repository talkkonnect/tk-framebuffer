package main

import (
	"fmt"
	"image"
	"image/draw"
)

// ChannelTreeNode is one row in the server channel tree (inactive rows keep depth).
type ChannelTreeNode struct {
	Name      string
	Depth     int
	UserCount int
	Active    bool
}

func channelDisplayName(name string) string {
	if name == "" {
		return "—"
	}
	return name
}

func inactiveChannelLabel(n ChannelTreeNode) string {
	return fmt.Sprintf("%s (%d)", channelDisplayName(n.Name), n.UserCount)
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

func drawChannelTree(img draw.Image, panel image.Rectangle, nodes []ChannelTreeNode) {
	x := panel.Min.X + 8
	y := panel.Min.Y + 30
	maxY := panel.Max.Y - 8
	const (
		activeRowH   = 16
		inactiveRowH = 18
	)

	active, rest := partitionChannelTree(nodes)

	if active != nil {
		if y > maxY {
			return
		}
		drawText(img, x, y+13, channelDisplayName(active.Name), colWhite, sizeChannelActive)
		y += activeRowH
	}

	for _, n := range rest {
		if y > maxY {
			drawText(img, panel.Max.X-24, maxY, "▼", colGreyText, sizeSmall)
			return
		}
		indent := x + n.Depth*14
		drawText(img, indent, y+14, inactiveChannelLabel(n), colGreyText, sizeSmall)
		y += inactiveRowH
	}
}
