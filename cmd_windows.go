//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func addHideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
