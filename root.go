// +build !windows

package main

import (
	"log"
	"os/user"
)

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return currentUser.Uid == "0"
}

