package main

// Pre-baked RGB565 sprites for static UI elements. Rebuilt once from the loaded theme palette.

const (
	userIconW = 7
	userIconH = 10
)

var (
	assetUserIconGreen  []byte
	assetUserIconRed    []byte
	assetUserIconPanel  []byte
	assetCommLEDOff     []byte
	assetCommLEDGreen   []byte
	assetCommLEDRed     []byte
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

func initAssetTiles(cfg *UIConfig) {
	pal := cfg.Palette
	assetUserIconGreen = rgb565SolidSprite(userIconW, userIconH, rgb565FromRGBA(pal.Green))
	assetUserIconRed = rgb565SolidSprite(userIconW, userIconH, rgb565FromRGBA(pal.Red))
	assetUserIconPanel = rgb565SolidSprite(userIconW, userIconH, rgb565FromRGBA(pal.Panel))
	assetCommLEDOff = rgb565SolidSprite(commLEDSize, commLEDSize, rgb565FromRGBA(pal.VUDim))
	assetCommLEDGreen = rgb565SolidSprite(commLEDSize, commLEDSize, rgb565FromRGBA(pal.Green))
	assetCommLEDRed = rgb565SolidSprite(commLEDSize, commLEDSize, rgb565FromRGBA(pal.Red))

	tileUserIconGreen = rgbaSolidTile(userIconW, userIconH, pal.Green)
	tileUserIconRed = rgbaSolidTile(userIconW, userIconH, pal.Red)
	tileUserIconPanel = rgbaSolidTile(userIconW, userIconH, pal.Panel)
	tileCommLEDOff = rgbaSolidTile(commLEDSize, commLEDSize, pal.VUDim)
	tileCommLEDGreen = rgbaSolidTile(commLEDSize, commLEDSize, pal.Green)
	tileCommLEDRed = rgbaSolidTile(commLEDSize, commLEDSize, pal.Red)
}
