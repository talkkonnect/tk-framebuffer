package main

import (
	"fmt"
	"image"
	"image/draw"
	"strings"
)

const messagePanelLineHeight = 13

func drawMumbleMessagePanel(img draw.Image, r image.Rectangle, st DisplayState, cfg *UIConfig) {
	if st.Transmitting {
		return
	}
	text := strings.TrimSpace(st.LastMessageText)
	if text == "" {
		return
	}

	pal := cfg.Palette
	caps := cfg.Captions

	fillRect(img, r, pal.VUDim)
	strokeRect(img, r, pal.PanelEdge, 1)

	sender := strings.TrimSpace(st.LastMessageSender)
	if sender == "" {
		sender = caps.MessageFromServerLabel
	}
	header := fmt.Sprintf(caps.MessageFromFormat, sender)

	x := r.Min.X + 6
	y := r.Min.Y + 12
	drawText(img, x, y, header, pal.Orange, sizeSmall)

	maxWidth := r.Dx() - 12
	lines := wrapTextLines(text, maxWidth, sizeSmall)
	maxLines := (r.Dy() - 22) / messagePanelLineHeight
	if maxLines < 1 {
		return
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[maxLines-1] = truncateTextWithEllipsis(lines[maxLines-1], maxWidth, sizeSmall)
	}

	y += 6
	for _, line := range lines {
		y += messagePanelLineHeight
		if y > r.Max.Y-2 {
			break
		}
		drawText(img, x, y, line, pal.White, sizeSmall)
	}
}

func truncateTextWithEllipsis(text string, maxWidth int, size fontSize) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	ellipsis := "..."
	if measureTextWidth(text, size) <= maxWidth {
		return text
	}
	runes := []rune(text)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + ellipsis
		if measureTextWidth(candidate, size) <= maxWidth {
			return candidate
		}
	}
	return ellipsis
}
