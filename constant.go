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
	WorkDir = filepath.Join(filepath.Dir(executable), "prog")

	err = os.MkdirAll(WorkDir, 0755)
	if err != nil {
		panic(err)
	}

	ConfigFile = filepath.Join(WorkDir, "config.json")

	LogFile = filepath.Join(WorkDir, "xx-"+time.Now().Format(time.DateOnly)+".log")

	PersistFile = filepath.Join(WorkDir, "persist.json")

	switch runtime.GOOS {
	case "windows":
		BinaryFile = filepath.Join(WorkDir, "sing-box.exe")
	default:
		BinaryFile = filepath.Join(WorkDir, "sing-box")
	}
}
