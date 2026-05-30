package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"time"
)

func main() {
	talkkonnectURL := flag.String("talkkonnect", "http://127.0.0.1:8080/uistatus", "talkkonnect /uistatus JSON endpoint")
	logoPath := flag.String("logo", "talkkonnect-logo.png", "talkkonnect brand logo PNG for footer")
	mockMode := flag.Bool("mock", false, "run with demo data instead of talkkonnect")
	fbDevice := flag.String("fb", "/dev/fb0", "Linux framebuffer device")
	fontPath := flag.String("font", "DejaVuSans.ttf", "Latin font (DejaVu Sans)")
	thaiFontPath := flag.String("thai-font", defaultThaiFontPath(), "Thai font (Noto Sans Thai)")
	vt := flag.Int("vt", 1, "virtual terminal attached to the display")
	flag.Parse()

	fb, err := openLinuxFramebuffer(*fbDevice)
	if err != nil {
		fmt.Printf("Error opening display: %v\n", err)
		return
	}
	defer fb.close()

	releaseDisplay, err := acquireLinuxDisplay(fb, *vt)
	if err != nil {
		fmt.Printf("Warning: could not hide Linux console: %v\n", err)
		fmt.Println("Try: sudo sh -c 'echo 0 > /sys/class/vtconsole/vtcon1/bind'")
	} else {
		defer releaseDisplay()
	}

	initFonts(*fontPath, *thaiFontPath)
	initBrandLogo(*logoPath)

	var tk *talkkonnectClient
	if !*mockMode {
		tk = newTalkkonnectClient(*talkkonnectURL)
	}

	frame := image.NewRGBA(image.Rect(0, 0, fb.width, fb.height))

	fmt.Printf("talKKonnect framebuffer UI active on %s (%dx%d, stride=%d). Press Ctrl+C to stop.\n",
		*fbDevice, fb.width, fb.height, fb.stride)
	if *mockMode {
		fmt.Println("Running in mock mode (-mock).")
	} else {
		fmt.Printf("Polling talkkonnect at %s\n", *talkkonnectURL)
	}

	// Single frame ticker; stopped on exit via defer (no extra Timer/Ticker instances).
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var lastErr time.Time
	var elapsed elapsedTracker
	frameNum := 0

	for now := range ticker.C {
		display := mockDisplayState()
		if tk != nil {
			st, err := tk.fetch()
			if err != nil {
				if time.Since(lastErr) > 5*time.Second {
					fmt.Printf("talkkonnect status error: %v\n", err)
					lastErr = time.Now()
				}
				display = offlineDisplayState()
			} else {
				display = st.toDisplayState()
			}
		}

		display.Elapsed = elapsed.update(now, display.Transmitting, display.Receiving)

		renderFrame(frame, fb.width, fb.height, display, signalLevel(display), now)
		if err := fb.blitRGBA(frame); err != nil {
			fmt.Printf("framebuffer blit error: %v\n", err)
		}

		frameNum++
		if frameNum == 1 {
			fmt.Println("First frame drawn to screen.")
		}
	}
}

func defaultThaiFontPath() string {
	candidates := []string{
		"NotoSansThai-Regular.ttf",
		"/usr/share/fonts/truetype/noto/NotoSansThai-Regular.ttf",
		"/usr/share/fonts/opentype/noto/NotoSansThai-Regular.ttf",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "NotoSansThai-Regular.ttf"
}
