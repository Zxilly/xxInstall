//go:build !windows

package main

import (
	"os/exec"
)

func addHideWindow(cmd *exec.Cmd) {
	// do nothing
}
