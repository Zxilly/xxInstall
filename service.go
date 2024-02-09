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

var runningCmd *exec.Cmd

type program struct{}

//goland:noinspection GoUnhandledErrorResult
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

	pe := persist{}
	err = pe.load()
	if err != nil {
		return err
	}

	if pe.PreferSystemInstall {
		BinaryFile, err = exec.LookPath("sing-box")
		if err != nil {
			_, lErr := logWriter.WriteString("Error finding system binary file: " + err.Error() + "\n")
			if lErr != nil {
				return err
			}
			return err
		}
	}

	go func() {
		logWriter.WriteString("Starting service...\n")

		runningCmd = exec.Command(BinaryFile, "run", "-c", absConfig, "-D", absWorkDir, "--disable-color")

		runningCmd.Stdout = logWriter
		runningCmd.Stderr = logWriter

		addHideWindow(runningCmd)
		logWriter.WriteString("Starting process...\n")
		err = runningCmd.Start()
		logWriter.WriteString("Process started...\n")
		if err != nil {
			logger.Infof("Error starting process: %s", err)
			return
		}
		logWriter.WriteString("Waiting process...\n")
		err = runningCmd.Wait()
		logWriter.WriteString("Process finished...\n")
		if err != nil {
			logger.Infof("Error waiting process: %s", err)
			return
		}
		err = s.Stop()
		if err != nil {
			logWriter.WriteString("Error stopping service: " + err.Error() + "\n")
			return
		}
	}()

	return nil
}

func (*program) Stop(s service.Service) error {
	if runningCmd != nil {
		if runtime.GOOS == "windows" {
			return runningCmd.Process.Kill()
		} else {
			err := runningCmd.Process.Signal(os.Interrupt)
			if err != nil {
				return err
			}
			go func() {
				<-time.After(5 * time.Second)
				err := runningCmd.Process.Kill()
				if err != nil {
					log.Println("Error killing process: ", err)
					return
				}
			}()

			state, err := runningCmd.Process.Wait()
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

	var dependencies []string
	switch runtime.GOOS {
	case "windows":
		dependencies = []string{
			"LanmanServer",
		}
	case "linux":
		dependencies = []string{
			"Wants=network-online.target",
			"After=network-online.target",
		}
	default:
		dependencies = []string{}
	}

	srv, err = service.New(&program{}, &service.Config{
		Name:        "XX Service",
		DisplayName: "XX Service",
		Description: "Service for XX",
		Arguments: []string{
			"service",
		},
		Dependencies: dependencies,
	})

	if err != nil {
		panic(err)
	}
}
