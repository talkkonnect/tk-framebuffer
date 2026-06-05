package main

// Pre-baked RGB565 sprites for static UI elements. Pixel format and dimensions are fixed at
// package init from compile-time palette constants — no rgb565() calls during rendering.

const (
	userIconW = 7
	userIconH = 10
)

var (
	assetUserIconGreen  = rgb565SolidSprite(userIconW, userIconH, pixGreen)
	assetUserIconRed    = rgb565SolidSprite(userIconW, userIconH, pixRed)
	assetUserIconPanel  = rgb565SolidSprite(userIconW, userIconH, pixPanel)
	assetCommLEDOff     = rgb565SolidSprite(commLEDSize, commLEDSize, pixVUDim)
	assetCommLEDGreen   = rgb565SolidSprite(commLEDSize, commLEDSize, pixGreen)
	assetCommLEDRed     = rgb565SolidSprite(commLEDSize, commLEDSize, pixRed)
)

// Pre-baked RGBA tiles for the software render buffer (same assets, RGBA layout).
var (
	tileUserIconGreen []byte
	tileUserIconRed   []byte
	tileUserIconPanel []byte
	tileCommLEDOff    []byte
	tileCommLEDGreen  []byte
	tileCommLEDRed    []byte
)

func init() {
	tileUserIconGreen = rgbaSolidTile(userIconW, userIconH, colGreen)
	tileUserIconRed = rgbaSolidTile(userIconW, userIconH, colRed)
	tileUserIconPanel = rgbaSolidTile(userIconW, userIconH, colPanel)
	tileCommLEDOff = rgbaSolidTile(commLEDSize, commLEDSize, colVUDim)
	tileCommLEDGreen = rgbaSolidTile(commLEDSize, commLEDSize, colGreen)
	tileCommLEDRed = rgbaSolidTile(commLEDSize, commLEDSize, colRed)
}
