package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type fontSize int

const (
	sizeLarge fontSize = iota
	sizeClock
	sizeStation
	sizeTitle
	sizeBody
	sizeLabel
	sizeSmall
	sizeChannelActive // 13pt: 50% of sizeLarge (26pt) for pinned active channel
)

type fontSet struct {
	large         font.Face
	clock         font.Face
	station       font.Face
	title         font.Face
	body          font.Face
	label         font.Face
	small         font.Face
	channelActive font.Face
}

var (
	fonts     fontSet
	thaiFonts fontSet
	thaiOK    bool
)

func isThaiRune(r rune) bool {
	return unicode.In(r, unicode.Thai)
}

func stringNeedsThai(s string) bool {
	for _, r := range s {
		if isThaiRune(r) {
			return true
		}
	}
	return false
}

func readFontFile(path string) ([]byte, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(fontBytes) < 4 || bytes.HasPrefix(fontBytes, []byte("<!")) || bytes.HasPrefix(fontBytes, []byte("<html")) {
		return nil, fmt.Errorf("not a valid font file (got HTML or corrupt data)")
	}
	return fontBytes, nil
}

func loadFontFaces(path string) (fontSet, error) {
	fontBytes, err := readFontFile(path)
	if err != nil {
		return fontSet{}, err
	}
	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return fontSet{}, fmt.Errorf("parse font: %w", err)
	}
	mk := func(size float64) (font.Face, error) {
		return opentype.NewFace(f, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
	}
	var fs fontSet
	var e error
	if fs.large, e = mk(26); e != nil {
		return fontSet{}, e
	}
	if fs.clock, e = mk(28); e != nil {
		return fontSet{}, e
	}
	if fs.station, e = mk(18); e != nil {
		return fontSet{}, e
	}
	if fs.title, e = mk(16); e != nil {
		return fontSet{}, e
	}
	if fs.body, e = mk(14); e != nil {
		return fontSet{}, e
	}
	if fs.label, e = mk(12); e != nil {
		return fontSet{}, e
	}
	if fs.small, e = mk(10); e != nil {
		return fontSet{}, e
	}
	if fs.channelActive, e = mk(13); e != nil {
		return fontSet{}, e
	}
	return fs, nil
}

func initFonts(latinPath, thaiPath string) {
	var err error
	fonts, err = loadFontFaces(latinPath)
	if err != nil {
		fmt.Printf("Latin font error (%s): %v\n", latinPath, err)
		os.Exit(1)
	}

	thaiOK = false
	if thaiPath == "" {
		return
	}
	if _, err := os.Stat(thaiPath); err != nil {
		fmt.Printf("Thai font not found (%s): %v — Thai text may not display\n", thaiPath, err)
		return
	}
	thaiFonts, err = loadFontFaces(thaiPath)
	if err != nil {
		fmt.Printf("Thai font error (%s): %v — Thai text may not display\n", thaiPath, err)
		return
	}
	thaiOK = true
	fmt.Printf("Thai language support enabled (%s)\n", thaiPath)
}

func faceFor(size fontSize, thai bool) font.Face {
	if thai && thaiOK {
		switch size {
		case sizeLarge:
			return thaiFonts.large
		case sizeClock:
			return thaiFonts.clock
		case sizeStation:
			return thaiFonts.station
		case sizeTitle:
			return thaiFonts.title
		case sizeBody:
			return thaiFonts.body
		case sizeLabel:
			return thaiFonts.label
		case sizeSmall:
			return thaiFonts.small
		case sizeChannelActive:
			return thaiFonts.channelActive
		}
	}
	switch size {
	case sizeLarge:
		return fonts.large
	case sizeClock:
		return fonts.clock
	case sizeStation:
		return fonts.station
	case sizeTitle:
		return fonts.title
	case sizeBody:
		return fonts.body
	case sizeLabel:
		return fonts.label
	case sizeSmall:
		return fonts.small
	case sizeChannelActive:
		return fonts.channelActive
	default:
		return fonts.body
	}
}

// drawText renders text with Latin font; switches to Thai font per rune when needed.
func drawText(img draw.Image, x, y int, text string, col color.Color, size fontSize) {
	if text == "" {
		return
	}

	d := &font.Drawer{
		Dst: img,
		Src: image.NewUniform(col),
		Dot: fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}

	if !thaiOK || !stringNeedsThai(text) {
		d.Face = faceFor(size, false)
		d.DrawString(text)
		return
	}

	for _, r := range text {
		d.Face = faceFor(size, isThaiRune(r))
		d.DrawString(string(r))
	}
}

// drawTextCentered places a single-line label in the horizontal and vertical center of r.
func drawTextCentered(img draw.Image, r image.Rectangle, text string, col color.Color, size fontSize) {
	if text == "" {
		return
	}
	f := faceFor(size, stringNeedsThai(text))
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: f,
	}
	advance := font.MeasureString(f, text)
	bounds, _ := d.BoundString(text)
	textW := advance.Ceil()
	textH := (bounds.Max.Y - bounds.Min.Y).Ceil()
	x := r.Min.X + (r.Dx()-textW)/2
	y := r.Min.Y + (r.Dy()-textH)/2 - bounds.Min.Y.Ceil()
	d.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	d.DrawString(text)
}
