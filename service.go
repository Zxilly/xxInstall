package main

import (
	"github.com/kardianos/service"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var cmd *exec.Cmd

type program struct{}

func (*program) Start(s service.Service) error {
	logger, err := s.Logger(nil)
	logWriter, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %s", err)
	}

	if err != nil {
		return err
	}

	absConfig, err := filepath.Abs(ConfigFile)
	if err != nil {
		_, err = logWriter.WriteString("Error getting absolute path of config file: " + err.Error() + "\n")
		if err != nil {
			return err
		}
		return err
	}

	absWorkDir, err := filepath.Abs(WorkDir)
	if err != nil {
		_, err = logWriter.WriteString("Error getting absolute path of work dir: " + err.Error() + "\n")
		if err != nil {
			return err
		}
		return err
	}

	go func() {
		logWriter.WriteString("Starting service...\n")

		cmd = exec.Command(BinaryFile, "run", "-c", absConfig, "-D", absWorkDir, "--disable-color")

		cmd.Stdout = logWriter
		cmd.Stderr = logWriter

		addHideWindow(cmd)
		logWriter.WriteString("Starting process...\n")
		err = cmd.Start()
		logWriter.WriteString("Process started...\n")
		if err != nil {
			logger.Infof("Error starting process: %s", err)
			return
		}
		logWriter.WriteString("Waiting process...\n")
		err = cmd.Wait()
		logWriter.WriteString("Process finished...\n")
		if err != nil {
			logger.Infof("Error waiting process: %s", err)
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
