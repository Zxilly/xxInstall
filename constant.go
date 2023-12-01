package main

import (
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var WorkDir string

var BinaryFile string

var ConfigFile string
var LogFile string
var PersistFile string

func init() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	WorkDir = filepath.Dir(executable) + string(os.PathSeparator) + "prog"

	err = os.MkdirAll(WorkDir, 0755)
	if err != nil {
		panic(err)
	}

	ConfigFile = WorkDir + string(os.PathSeparator) + "config.json"

	LogFile = WorkDir + string(os.PathSeparator) + "xx-" + time.Now().Format(time.DateOnly) + ".log"

	PersistFile = WorkDir + string(os.PathSeparator) + "persist.json"

	if runtime.GOOS == "windows" {
		BinaryFile = WorkDir + string(os.PathSeparator) + "sing-box.exe"
	} else {
		BinaryFile = WorkDir + string(os.PathSeparator) + "sing-box"
	}
}
