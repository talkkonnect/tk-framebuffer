package main

import "image/color"

func rgb565(r, g, b byte) uint16 {
	return uint16(r>>3)<<11 | uint16(g>>2)<<5 | uint16(b>>3)
}

func rgb565FromRGBA(c color.RGBA) uint16 {
	return rgb565(c.R, c.G, c.B)
}

// Palette RGB565 values pre-computed at compile time (no runtime bit packing on hot paths).
const (
	pixBlack       uint16 = 0x0000
	pixBackground  uint16 = 0x0862
	pixPanel       uint16 = 0x18C3
	pixPanelHead   uint16 = 0x2146
	pixPanelEdge   uint16 = 0x3A4B
	pixBlue        uint16 = 0x4C38
	pixBlueDim     uint16 = 0x2A2C
	pixGreyText    uint16 = 0xAD76
	pixWhite       uint16 = 0xEF7E
	pixOrange      uint16 = 0xEBA4
	pixRed         uint16 = 0xD1A8
	pixGreen       uint16 = 0x3DEC
	pixVUDim       uint16 = 0x10A3
	pixVUGreen     uint16 = 0x3548
	pixVUYellow    uint16 = 0xD5E7
	pixLightYellow uint16 = 0xCCA7
	pixVURed       uint16 = 0xC9A6
)

func rgb565RowBytes(width int, pix uint16) []byte {
	row := make([]byte, width*2)
	lo := byte(pix & 0xff)
	hi := byte(pix >> 8)
	for i := 0; i < len(row); i += 2 {
		row[i] = lo
		row[i+1] = hi
	}
	return row
}

// rgb565SolidSprite returns w×h RGB565 pixels as a contiguous byte slice (row-major, 2 bytes/pixel).
func rgb565SolidSprite(w, h int, pix uint16) []byte {
	if w <= 0 || h <= 0 {
		return nil
	}
	row := rgb565RowBytes(w, pix)
	out := make([]byte, len(row)*h)
	for y := 0; y < h; y++ {
		copy(out[y*len(row):], row)
	}
	return out
}

func rgbaSolidRow(width int, c color.RGBA) []byte {
	row := make([]byte, width*4)
	for i := 0; i < len(row); i += 4 {
		row[i] = c.R
		row[i+1] = c.G
		row[i+2] = c.B
		row[i+3] = c.A
	}
	return row
}

// rgbaSolidTile returns a pre-filled RGBA image tile for fast blitting of solid rectangles.
func rgbaSolidTile(w, h int, c color.RGBA) []byte {
	if w <= 0 || h <= 0 {
		return nil
	}
	row := rgbaSolidRow(w, c)
	out := make([]byte, len(row)*h)
	for y := 0; y < h; y++ {
		copy(out[y*len(row):], row)
	}
	return out
}
