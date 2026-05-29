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
	file   *os.File
	mem    []byte
	width  int
	height int
	bpp    int
	stride int
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

func rgb565(r, g, b byte) uint16 {
	return uint16(r>>3)<<11 | uint16(g>>2)<<5 | uint16(b>>3)
}

func (fb *linuxFramebuffer) fillSolid(c color.RGBA) {
	if fb.bpp != 2 {
		return
	}
	pix := rgb565(c.R, c.G, c.B)
	lo := byte(pix & 0xff)
	hi := byte(pix >> 8)
	for y := 0; y < fb.height; y++ {
		row := fb.mem[y*fb.stride : y*fb.stride+fb.width*2]
		for x := 0; x < fb.width; x++ {
			i := x * 2
			row[i] = lo
			row[i+1] = hi
		}
	}
}

func (fb *linuxFramebuffer) blitRGBA(img *image.RGBA) error {
	if fb.bpp != 2 {
		return fmt.Errorf("unsupported framebuffer bpp: %d", fb.bpp*8)
	}
	for y := 0; y < fb.height; y++ {
		row := fb.mem[y*fb.stride : y*fb.stride+fb.width*2]
		for x := 0; x < fb.width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pix := rgb565(byte(r>>8), byte(g>>8), byte(b>>8))
			i := x * 2
			row[i] = byte(pix & 0xff)
			row[i+1] = byte(pix >> 8)
		}
	}
	return nil
}
