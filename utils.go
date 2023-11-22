package main

import (
	"os"
	"runtime"
)

func isExist(path string) bool {
	_, err := os.Stat(WorkDir + string(os.PathSeparator) + path)
	return err == nil || os.IsExist(err)
}

func isExecutableExist(path string) bool {
	if runtime.GOOS == "windows" {
		return isExist(path + ".exe")
	} else {
		return isExist(path)
	}
}
