//go:build linux

package main

import (
	"fmt"
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

func framebufferConsoleBindPath() string {
	entries, err := os.ReadDir("/sys/class/vtconsole")
	if err != nil {
		return "/sys/class/vtconsole/vtcon1/bind"
	}
	for _, entry := range entries {
		nameFile := filepath.Join("/sys/class/vtconsole", entry.Name(), "name")
		nameBytes, err := os.ReadFile(nameFile)
		if err != nil {
			continue
		}
		if strings.Contains(string(nameBytes), "frame buffer") {
			return filepath.Join("/sys/class/vtconsole", entry.Name(), "bind")
		}
	}
	return "/sys/class/vtconsole/vtcon1/bind"
}

func setFramebufferConsoleBound(path string, on bool) error {
	val := "0"
	if on {
		val = "1"
	}
	return os.WriteFile(path, []byte(val), 0)
}

func stopGetty(vt int) func() {
	unit := fmt.Sprintf("getty@tty%d.service", vt)
	wasActive := exec.Command("systemctl", "is-active", "--quiet", unit).Run() == nil
	if wasActive {
		_ = exec.Command("systemctl", "stop", unit).Run()
	}
	return func() {
		if wasActive {
			_ = exec.Command("systemctl", "start", unit).Run()
		}
	}
}

// acquireLinuxDisplay hides the login/console on the attached screen so framebuffer
// writes are visible. Returns a cleanup func that restores the console.
func acquireLinuxDisplay(fb *linuxFramebuffer, vt int) (func(), error) {
	if vt <= 0 {
		vt = 1
	}

	bindPath := framebufferConsoleBindPath()
	consoleWasBound := true
	if raw, err := os.ReadFile(bindPath); err == nil && strings.TrimSpace(string(raw)) == "0" {
		consoleWasBound = false
	}

	_ = exec.Command("chvt", fmt.Sprintf("%d", vt)).Run()
	restoreGetty := stopGetty(vt)

	ttyPath := fmt.Sprintf("/dev/tty%d", vt)
	if tty, err := os.OpenFile(ttyPath, os.O_RDWR, 0); err == nil {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, tty.Fd(), kdSetMode, kdGraphics)
		tty.Close()
	}

	if consoleWasBound {
		if err := setFramebufferConsoleBound(bindPath, false); err != nil {
			restoreGetty()
			return nil, fmt.Errorf("release framebuffer console: %w (run as root)", err)
		}
	}

	if fb.file != nil {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fb.file.Fd(), fbIOBlank, fbBlankUnblank)
	}

	// Paint immediately so the screen changes before font rendering starts.
	fb.fillSolid(colBackground)

	cleanup := func() {
		if consoleWasBound {
			_ = setFramebufferConsoleBound(bindPath, true)
		}
		if tty, err := os.OpenFile(ttyPath, os.O_RDWR, 0); err == nil {
			_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, tty.Fd(), kdSetMode, kdText)
			tty.Close()
		}
		restoreGetty()
	}

	return cleanup, nil
}
