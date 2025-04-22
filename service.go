package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2"
)

var runningCmd *exec.Cmd

type program struct{}

var exiting = atomic.Bool{}

//goland:noinspection GoUnhandledErrorResult
func (*program) Start(s service.Service) error {
	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatalf("Error opening log file: %s", err)
	}
	logWriter := &lumberjack.Logger{
		Filename:   LogFile,
		MaxSize:    20,
		MaxBackups: 3,
		MaxAge:     28,
	}
	log.SetOutput(logWriter)

	logBoth := func(format string, v ...interface{}) {
		log.Printf(format, v...)
		logger.Infof(format, v...)
	}

	absWorkDir, err := filepath.Abs(WorkDir)
	if err != nil {
		logBoth("Error getting absolute path of work directory: %s", err)
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
			logBoth("Error getting system installed binary: %s", err)
			return err
		}
	}

	logBoth("Starting service...")

	runningCmd = exec.Command(BinaryFile, "run", "-c", ConfigFile, "-D", absWorkDir, "--disable-color")

	runningCmd.Stdout = logWriter
	runningCmd.Stderr = logWriter

	addHideWindow(runningCmd)
	err = setupDNS()
	if err != nil {
		logBoth("Error setting up DNS: %s", err)
		return err
	}

	logBoth("Starting process...")
	err = runningCmd.Start()
	logBoth("Process started...")
	if err != nil {
		logBoth("Error starting process: %s", err)
		return err
	}

	go func() {
		logBoth("Waiting process...")
		err = runningCmd.Wait()
		logBoth("Process finished...")
		if err != nil {
			logger.Infof("Error waiting process: %s", err)
			return
		}
		if !exiting.Load() {
			err = restoreDNS()
			if err != nil {
				logBoth("Error restoring DNS: %s", err)
			}
			err = s.Stop()
			if err != nil {
				logBoth("Error stopping service: %s", err)
				return
			}
		}
	}()

	return nil
}

func (*program) Stop(s service.Service) error {
	exiting.Store(true)
	if runningCmd != nil {
		err := restoreDNS()
		if err != nil {
			log.Println("Error restoring DNS:", err)
		}

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
