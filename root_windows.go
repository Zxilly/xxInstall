//go:build windows

package main

import (
	"github.com/Microsoft/go-winio"
	"github.com/mitchellh/go-ps"
	"golang.org/x/sys/windows"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const pipeBase = `\\.\pipe\`

const logoutName = pipeBase + "XXLogOutput"

func requireRoot() {
	if !isRoot() {
		initServer()
		runMeElevated()
		runAsDisplay()
	} else {
		process, err := ps.FindProcess(os.Getppid())
		if err != nil {
			log.Fatalf("Error finding parent process: %s", err)
		}
		currentExe, err := os.Executable()
		if err != nil {
			log.Fatalf("Error getting current executable: %s", err)
		}
		currentExeBase := filepath.Base(currentExe)

		if process.Executable() == currentExeBase {
			// elevated process
			runAsSender()

			go func() {
				parent := os.Getppid()
				for {
					pp, err := ps.FindProcess(parent)
					if pp == nil && err == nil {
						// parent process is gone
						senderConn.Close()
						os.Exit(0)
					}
					time.Sleep(1 * time.Second)
				}
			}()
		}

		// do nothing
	}
}

var senderConn net.Conn
var shouldSend = false

func runAsSender() {
	conn, err := winio.DialPipe(logoutName, nil)
	if err != nil {
		log.Fatalf("Error connecting to stdout server: %s", err)
	}

	rootCmd.SetOut(conn)
	log.SetOutput(conn)

	senderConn = conn
	shouldSend = true
}

var listener net.Listener

func initServer() {
	var err error
	listener, err = winio.ListenPipe(logoutName, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
	if err != nil {
		log.Fatalf("Error creating stdout server: %s", err)
	}
}

func runAsDisplay() {
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatalf("Error closing stdout server: %s", err)
		}
	}(listener)

	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("Error accepting stdout connection: %s", err)
	}

	_, _ = io.Copy(os.Stdout, conn)

	os.Exit(0)
}

func isRoot() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

func runMeElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 0 //SW_HIDE

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		log.Fatalf("Error elevating: %s", err)
	}
}
