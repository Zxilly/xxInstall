package main

import (
	"os"
	"runtime"
)

var WorkDir = "prog"

var BinaryFile string

var ConfigFile = WorkDir + string(os.PathSeparator) + "config.json"

var LogFile = WorkDir + string(os.PathSeparator) + "xx.log"

var PersistFile = WorkDir + string(os.PathSeparator) + "persist.json"

func init() {
	err := os.MkdirAll(WorkDir, 0755)
	if err != nil {
		panic(err)
	}

	if runtime.GOOS == "windows" {
		BinaryFile = WorkDir + string(os.PathSeparator) + "sing-box.exe"
	} else {
		BinaryFile = WorkDir + string(os.PathSeparator) + "sing-box"
	}
}
