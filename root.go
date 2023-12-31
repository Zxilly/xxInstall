//go:build !windows
// +build !windows

package main

import (
	"log"
	"net"
	"os/user"
)

var senderConn net.Conn
var shouldSend = false

func requireRoot() {
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
