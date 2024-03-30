package main

import (
	"os"
	"path/filepath"
	"runtime"
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
	baseDir := filepath.Dir(executable)
	WorkDir = filepath.Join(baseDir, "prog")

	err = os.MkdirAll(WorkDir, 0755)
	if err != nil {
		panic(err)
	}

	//ConfigFile = WorkDir + string(os.PathSeparator) + "config.json"
	ConfigFile = filepath.Join(baseDir, "config", "config.json")

	LogFile = filepath.Join(baseDir, "logs", "log.txt")

	PersistFile = filepath.Join(baseDir, "config", "persist.json")

	if runtime.GOOS == "windows" {
		BinaryFile = filepath.Join(WorkDir, "sing-box.exe")
	} else {
		BinaryFile = filepath.Join(WorkDir, "sing-box")
	}
}
