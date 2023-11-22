//go:build windows

package main

import (
	"github.com/Microsoft/go-winio"
	"github.com/mitchellh/go-ps"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const pipeBase = `\\.\pipe\`

const logoutName = pipeBase + "XXLogOutput"

const closeType = 7

func init() {
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
			rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {

				return nil
			}
		}

		// do nothing
	}
}

type ipcWriter struct {
	conn net.Conn
}

func appendLog(s string) {
	f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %s", err)
	}
	defer f.Close()

	_, err = f.WriteString(s + "\n")
	if err != nil {
		log.Fatalf("Error writing to log file: %s", err)
	}
}

func runAsSender() {
	conn, err := winio.DialPipe(logoutName, nil)
	if err != nil {
		log.Fatalf("Error connecting to stdout server: %s", err)
	}

	rootCmd.SetOut(conn)
	log.SetOutput(conn)
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
	var sid *windows.SID

	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		log.Fatalf("SID Error: %s", err)
		return false
	}
	defer func(sid *windows.SID) {
		err := windows.FreeSid(sid)
		if err != nil {
			log.Fatalf("SID Free Error: %s", err)
		}
	}(sid)

	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		log.Fatalf("Token Membership Error: %s", err)
		return false
	}

	return member
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
