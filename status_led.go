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

func drawTalkkonnectStatusLED(img draw.Image, r image.Rectangle, talkkonnectOK bool, now time.Time) {
	ledX := r.Max.X - commLEDSize - 4
	ledY := r.Min.Y + (r.Dy()-commLEDSize)/2
	led := image.Rect(ledX, ledY, ledX+commLEDSize, ledY+commLEDSize)

	fillRect(img, led, colVUDim)
	strokeRect(img, led, colPanelEdge, 1)

	if !commLEDLit(now) {
		return
	}

	lit := colGreen
	if !talkkonnectOK {
		lit = colRed
	}
	fillRect(img, led, lit)
	strokeRect(img, led, colWhite, 1)
}
