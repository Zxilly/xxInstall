package main

import (
	"github.com/kardianos/service"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

var cmd *exec.Cmd

type program struct{}

func (*program) Start(s service.Service) error {
	absConfig, err := filepath.Abs(ConfigFile)
	if err != nil {
		return err
	}

	go func() {
		cmd = exec.Command(BinaryFile, "run", "-c", absConfig)

		logWriter, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error opening log file: %s", err)
		}
		cmd.Stdout = logWriter
		cmd.Stderr = logWriter

		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		}

		err = cmd.Start()
		if err != nil {
			log.Println("Error starting process: ", err)
			return
		}
	}()

	return nil
}

func (*program) Stop(s service.Service) error {
	if cmd != nil {
		if runtime.GOOS == "windows" {
			return cmd.Process.Kill()
		} else {
			err := cmd.Process.Signal(os.Interrupt)
			if err != nil {
				return err
			}
			go func() {
				<-time.After(5 * time.Second)
				err := cmd.Process.Kill()
				if err != nil {
					log.Println("Error killing process: ", err)
					return
				}
			}()

			state, err := cmd.Process.Wait()
			if err != nil {
				log.Println("Error waiting process: ", err)
				return err
			}
			if !state.Success() {
				return state.Sys().(error)
			}
		}
	}

	return nil
}

var _ service.Interface = (*program)(nil)

var srv service.Service

func init() {
	var err error
	srv, err = service.New(&program{}, &service.Config{
		Name:        "XX Service",
		DisplayName: "XX Service",
		Description: "Service for XX",
		Arguments: []string{
			"service",
		},
	})

	if err != nil {
		panic(err)
	}
}
