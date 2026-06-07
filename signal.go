package main

import (
	"image"
	"image/color"
	"image/draw"
)

const (
	signalMaxBarH = 10
	signalMinBarW = 10
	signalMinGap  = 4
)

var signalBarFractions = [5]float64{0.20, 0.40, 0.60, 0.80, 1.00}

// signalLevel returns 0..5 active bars (full scale on TX or RX).
func signalLevel(st DisplayState) int {
	if st.Transmitting || st.Receiving {
		return 5
	}
	return 0
}

func signalBarColor(barIndex int, cfg *UIConfig) color.Color {
	return vuSegmentColor(signalBarFractions[barIndex], cfg)
}

func signalBarLayout(availableW int) (barW, gap int) {
	const bars = 5
	const gaps = 4
	gap = signalMinGap
	barW = (availableW - gaps*gap) / bars
	if barW < signalMinBarW {
		barW = signalMinBarW
	}
	used := bars*barW + gaps*gap
	if used < availableW {
		gap += (availableW - used) / gaps
	}
	return barW, gap
}

func drawSignalBars(img draw.Image, x, baselineY, availableW, activeBars int, cfg *UIConfig) {
	barW, gap := signalBarLayout(availableW)
	for i := 0; i < 5; i++ {
		barH := int(float64(signalMaxBarH) * signalBarFractions[i])
		if barH < 1 {
			barH = 1
		}
		barY := baselineY - barH + 1
		col := color.Color(cfg.Palette.SignalInactive)
		if i < activeBars {
			col = signalBarColor(i, cfg)
		}
		fillRect(img, image.Rect(x, barY, x+barW, baselineY+1), col)
		x += barW + gap
	}
}

// drawSignalMeter renders a compact ascending bar graph.
// Returns the Y coordinate for the next UI element below it.
func drawSignalMeter(img draw.Image, x, y, w, activeBars int, cfg *UIConfig) int {
	drawText(img, x, y+10, cfg.Captions.SignalLevelLabel, cfg.Palette.GreyText, sizeSmall)

	baselineY := y + 12 + signalMaxBarH - 1
	drawSignalBars(img, x, baselineY, w, activeBars, cfg)

	return baselineY + 7
}
