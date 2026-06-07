//go:build linux

package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	fbIOBlank      = 0x4611
	fbBlankUnblank = 0
	kdSetMode      = 0x4B3A
	kdGraphics     = 0x01
	kdText         = 0x00
)

const vtconsoleClassDir = "/sys/class/vtconsole"

// unbindFramebufferConsole disconnects the kernel framebuffer console via sysfs so
// direct framebuffer writes are visible. Returns a cleanup func that rebinds the
// console if it was previously bound.
func unbindFramebufferConsole() func() {
	entries, err := os.ReadDir(vtconsoleClassDir)
	if err != nil {
		return func() {}
	}

	var bindPath string
	for _, entry := range entries {
		nameFile := filepath.Join(vtconsoleClassDir, entry.Name(), "name")
		nameBytes, err := os.ReadFile(nameFile)
		if err != nil {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(string(nameBytes)))
		if !strings.Contains(strings.ReplaceAll(name, " ", ""), "framebuffer") {
			continue
		}
		bindPath = filepath.Join(vtconsoleClassDir, entry.Name(), "bind")
		break
	}
	if bindPath == "" {
		return func() {}
	}

	bindBytes, err := os.ReadFile(bindPath)
	if err != nil {
		return func() {}
	}
	if strings.TrimSpace(string(bindBytes)) != "1" {
		return func() {}
	}

	if err := os.WriteFile(bindPath, []byte("0"), 0); err != nil {
		return func() {}
	}

	return func() {
		_ = os.WriteFile(bindPath, []byte("1"), 0)
	}
}

// acquireLinuxDisplay hides the login/console on the attached screen so framebuffer
// writes are visible. Returns a cleanup func that restores the console.
func acquireLinuxDisplay(fb *linuxFramebuffer, vt int, bg color.RGBA) (func(), error) {
	if vt <= 0 {
		vt = 1
	}

	_ = exec.Command("chvt", fmt.Sprintf("%d", vt)).Run()
	restoreConsole := unbindFramebufferConsole()

	ttyPath := fmt.Sprintf("/dev/tty%d", vt)
	if tty, err := os.OpenFile(ttyPath, os.O_RDWR, 0); err == nil {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, tty.Fd(), kdSetMode, kdGraphics)
		tty.Close()
	}

	if fb.file != nil {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fb.file.Fd(), fbIOBlank, fbBlankUnblank)
	}

	// Paint immediately so the screen changes before font rendering starts.
	fb.fillSolid(bg)

	cleanup := func() {
		if tty, err := os.OpenFile(ttyPath, os.O_RDWR, 0); err == nil {
			_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, tty.Fd(), kdSetMode, kdText)
			tty.Close()
		}
		restoreConsole()
	}

	return cleanup, nil
}
