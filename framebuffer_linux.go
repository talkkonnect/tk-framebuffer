//go:build linux

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"syscall"
	"unsafe"
)

const (
	fbIOGetVarScreeninfo = 0x4600
	fbIOGetFixScreeninfo = 0x4602
)

type fbBitField struct {
	Offset, Length, MsbRight uint32
}

type fbVarScreeninfo struct {
	Xres, Yres, XresVirtual, YresVirtual, Xoffset, Yoffset, BitsPerPixel, Grayscale uint32
	Red, Green, Blue, Transp                                                        fbBitField
	Nonstd, Activate, Height, Width, AccelFlags                                     uint32
}

type fbFixScreeninfo struct {
	ID           [16]byte
	SmemStart    uintptr
	SmemLen      uint32
	Type         uint32
	TypeAux      uint32
	Visual       uint32
	Xpanstep     uint32
	Ypanstep     uint32
	Ywrapstep    uint32
	LineLength   uint32
	MMIOStart    uintptr
	MMIOStartLen uint32
	Accel        uint32
}

type linuxFramebuffer struct {
	file    *os.File
	mem     []byte
	width   int
	height  int
	bpp     int
	stride  int
	blitRow []byte
}

func openLinuxFramebuffer(path string) (*linuxFramebuffer, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var vinfo fbVarScreeninfo
	var finfo fbFixScreeninfo
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), fbIOGetVarScreeninfo, uintptr(unsafe.Pointer(&vinfo))); errno != 0 {
		file.Close()
		return nil, fmt.Errorf("FBIOGET_VSCREENINFO: %v", errno)
	}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), fbIOGetFixScreeninfo, uintptr(unsafe.Pointer(&finfo))); errno != 0 {
		file.Close()
		return nil, fmt.Errorf("FBIOGET_FSCREENINFO: %v", errno)
	}

	width, height := int(vinfo.Xres), int(vinfo.Yres)
	bpp := int(vinfo.BitsPerPixel / 8)
	if bpp <= 0 {
		bpp = 2
	}

	size := int(finfo.SmemLen)
	if size <= 0 {
		size = width * height * bpp
	}

	mem, err := syscall.Mmap(int(file.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("mmap framebuffer: %w", err)
	}

	stride := int(finfo.LineLength)
	if stride <= 0 {
		stride = width * bpp
	}

	return &linuxFramebuffer{
		file:   file,
		mem:    mem,
		width:  width,
		height: height,
		bpp:    bpp,
		stride: stride,
	}, nil
}

func (fb *linuxFramebuffer) close() error {
	if fb.mem != nil {
		_ = syscall.Munmap(fb.mem)
		fb.mem = nil
	}
	if fb.file != nil {
		err := fb.file.Close()
		fb.file = nil
		return err
	}
	return nil
}

func (fb *linuxFramebuffer) rowLen() int {
	return fb.width * 2
}

func (fb *linuxFramebuffer) ensureBlitRow() []byte {
	rowLen := fb.rowLen()
	if len(fb.blitRow) < rowLen {
		fb.blitRow = make([]byte, rowLen)
	}
	return fb.blitRow[:rowLen]
}

func rgb565Store(dst []byte, x int, pix uint16) {
	i := x * 2
	dst[i] = byte(pix & 0xff)
	dst[i+1] = byte(pix >> 8)
}

func rgbaRowToRGB565(dst, src []byte, width int) {
	for x := 0; x < width; x++ {
		i := x * 4
		pix := rgb565(src[i], src[i+1], src[i+2])
		rgb565Store(dst, x, pix)
	}
}

func (fb *linuxFramebuffer) fillSolid(c color.RGBA) {
	if fb.bpp != 2 {
		return
	}
	rowLen := fb.rowLen()
	if rowLen == 0 || fb.height == 0 {
		return
	}

	pix := rgb565FromRGBA(c)
	first := fb.mem[:rowLen]
	lo := byte(pix & 0xff)
	hi := byte(pix >> 8)
	for i := 0; i < rowLen; i += 2 {
		first[i] = lo
		first[i+1] = hi
	}
	for y := 1; y < fb.height; y++ {
		copy(fb.mem[y*fb.stride:y*fb.stride+rowLen], first)
	}
}

// blitRGB565 copies a pre-baked w×h RGB565 sprite into framebuffer memory at (x, y).
func (fb *linuxFramebuffer) blitRGB565(x, y, w, h int, sprite []byte) {
	if fb.bpp != 2 || w <= 0 || h <= 0 {
		return
	}
	rowBytes := w * 2
	if len(sprite) < rowBytes*h {
		return
	}
	for row := 0; row < h; row++ {
		if y+row < 0 || y+row >= fb.height {
			continue
		}
		dstY := y + row
		dstStart := dstY*fb.stride + x*2
		dstEnd := dstStart + rowBytes
		if x < 0 || dstStart >= len(fb.mem) {
			continue
		}
		if dstEnd > dstY*fb.stride+fb.width*2 {
			dstEnd = dstY*fb.stride + fb.width*2
		}
		if dstEnd <= dstStart {
			continue
		}
		n := dstEnd - dstStart
		copy(fb.mem[dstStart:dstEnd], sprite[row*rowBytes:row*rowBytes+n])
	}
}

func (fb *linuxFramebuffer) blitRGBA(img *image.RGBA) error {
	if fb.bpp != 2 {
		return fmt.Errorf("unsupported framebuffer bpp: %d", fb.bpp*8)
	}
	rowLen := fb.rowLen()
	if rowLen == 0 {
		return nil
	}

	rowBuf := fb.ensureBlitRow()
	bounds := img.Bounds()
	srcStride := img.Stride
	srcPix := img.Pix
	minX := bounds.Min.X

	for y := 0; y < fb.height; y++ {
		srcOff := (y-bounds.Min.Y)*srcStride + minX*4
		rgbaRowToRGB565(rowBuf, srcPix[srcOff:srcOff+rowLen*2], fb.width)
		copy(fb.mem[y*fb.stride:y*fb.stride+rowLen], rowBuf)
	}
	return nil
}
