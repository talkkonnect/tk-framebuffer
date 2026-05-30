package main

import (
	"fmt"
	"image"
	"image/draw"
	"os"

	xdraw "golang.org/x/image/draw"
	_ "image/png"
)

var brandLogo image.Image

func initBrandLogo(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Brand logo not found (%s): %v — footer will show placeholder\n", path, err)
		return
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Brand logo decode error (%s): %v\n", path, err)
		return
	}
	brandLogo = img
	fmt.Printf("Brand logo loaded (%s)\n", path)
}

func drawBrandLogo(img draw.Image, r image.Rectangle) {
	if brandLogo == nil {
		drawText(img, r.Min.X, r.Max.Y-8, "[logo]", colGreyText, sizeLabel)
		return
	}
	b := brandLogo.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 {
		return
	}
	scale := float64(r.Dy()) / float64(b.Dy())
	targetW := int(float64(b.Dx()) * scale)
	targetH := r.Dy()
	if targetW > r.Dx() {
		scale = float64(r.Dx()) / float64(b.Dx())
		targetW = r.Dx()
		targetH = int(float64(b.Dy()) * scale)
	}
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), brandLogo, b, xdraw.Over, nil)
	offset := image.Point{
		X: r.Min.X,
		Y: r.Min.Y + (r.Dy()-targetH)/2,
	}
	draw.Draw(img, image.Rect(offset.X, offset.Y, offset.X+targetW, offset.Y+targetH), dst, image.Point{}, draw.Over)
}
