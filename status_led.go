package main

import (
	"image"
	"image/draw"
	"time"
)

const (
	commLEDPeriod = 2 * time.Second
	commLEDOnTime = 500 * time.Millisecond
	commLEDSize   = 12
)

// commLEDLit is true during the 0.5s on-phase of each 2s blink cycle.
func commLEDLit(now time.Time) bool {
	phase := now.UnixNano() % int64(commLEDPeriod)
	return time.Duration(phase) < commLEDOnTime
}

func drawTalkkonnectStatusLED(img draw.Image, r image.Rectangle, talkkonnectOK bool, now time.Time, cfg *UIConfig) {
	pal := cfg.Palette
	ledX := r.Max.X - commLEDSize - 4
	ledY := r.Min.Y + (r.Dy()-commLEDSize)/2

	if rgba, ok := img.(*image.RGBA); ok {
		blitRGBATile(rgba, ledX, ledY, commLEDSize, commLEDSize, tileCommLEDOff)
		strokeRect(img, image.Rect(ledX, ledY, ledX+commLEDSize, ledY+commLEDSize), pal.PanelEdge, 1)
		if !commLEDLit(now) {
			return
		}
		lit := tileCommLEDGreen
		if !talkkonnectOK {
			lit = tileCommLEDRed
		}
		blitRGBATile(rgba, ledX, ledY, commLEDSize, commLEDSize, lit)
		strokeRect(img, image.Rect(ledX, ledY, ledX+commLEDSize, ledY+commLEDSize), pal.White, 1)
		return
	}

	led := image.Rect(ledX, ledY, ledX+commLEDSize, ledY+commLEDSize)
	fillRect(img, led, pal.VUDim)
	strokeRect(img, led, pal.PanelEdge, 1)
	if !commLEDLit(now) {
		return
	}
	lit := pal.Green
	if !talkkonnectOK {
		lit = pal.Red
	}
	fillRect(img, led, lit)
	strokeRect(img, led, pal.White, 1)
}
