//go:build !windows
// +build !windows

package main

import (
	"log"
	"os/user"
)

var senderConn net.Conn
var shouldSend = false

func init() {
	if !isRoot() {
		log.Fatalf("Please run as root.")
	}
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return currentUser.Uid == "0"
}
