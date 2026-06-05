package main

import (
	"image"
	"image/color"
	"image/draw"
)

var fillRowBuf []byte

func fillRect(img draw.Image, r image.Rectangle, col color.Color) {
	if rgba, ok := img.(*image.RGBA); ok {
		c := color.RGBAModel.Convert(col).(color.RGBA)
		fillRectRGBA(rgba, r, c)
		return
	}
	draw.Draw(img, r, &image.Uniform{col}, image.Point{}, draw.Src)
}

func fillRectRGBA(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	w := r.Dx()
	h := r.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	rowLen := w * 4
	if len(fillRowBuf) < rowLen {
		fillRowBuf = make([]byte, rowLen)
	}
	row := fillRowBuf[:rowLen]
	for i := 0; i < rowLen; i += 4 {
		row[i] = c.R
		row[i+1] = c.G
		row[i+2] = c.B
		row[i+3] = c.A
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		off := y*img.Stride + r.Min.X*4
		copy(img.Pix[off:off+rowLen], row)
	}
}

func blitRGBATile(img *image.RGBA, x, y, w, h int, tile []byte) {
	if w <= 0 || h <= 0 || len(tile) < w*4*h {
		return
	}
	rowLen := w * 4
	for row := 0; row < h; row++ {
		dstOff := (y+row)*img.Stride + x*4
		srcOff := row * rowLen
		copy(img.Pix[dstOff:dstOff+rowLen], tile[srcOff:srcOff+rowLen])
	}
}
