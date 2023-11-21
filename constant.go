package main

import (
	"os"
	"runtime"
)

var WORK_DIR = "prog"

var BINARY_FILE string

var CONFIG_FILE = WORK_DIR + string(os.PathSeparator) + "config.json"

func init() {
	os.MkdirAll(WORK_DIR, 0755)

	if runtime.GOOS == "windows" {
		BINARY_FILE = WORK_DIR + string(os.PathSeparator) + "sing-box.exe"
	} else {
		BINARY_FILE = WORK_DIR + string(os.PathSeparator) + "sing-box"
	}
}