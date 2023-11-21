package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

func init() {
	if !isRoot() {
		log.Fatal("Please run as root.")
	}
}

func startCmdRun(cmd *cobra.Command, args []string) {
	err := srv.Start()
	if err != nil {
		log.Fatalf("Error starting service: %s", err)
	} else {
		log.Println("Service started.")
	}
}

func stopCmdRun(cmd *cobra.Command, args []string) {
	err := srv.Stop()
	if err != nil {
		log.Fatalf("Error stopping service: %s", err)
	} else {
		log.Println("Service stopped.")
	}
}

func restartCmdRun(cmd *cobra.Command, args []string) {
	err := srv.Restart()
	if err != nil {
		log.Fatalf("Error restarting service: %s", err)
	} else {
		log.Println("Service restarted.")
	}
}

func installCmdRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Println("Please specify an url for install.")
		return
	}
	u := args[0]

	// test the url is valid
	url, err := url.Parse(u)
	if err != nil {
		log.Fatalf("Error parsing url: %s", err)
	}

	downloadConfig(url.String())
	if !isExecutableExist("sing-box") {
		downloadBinary()
	}

	p := persist{ConfigURL: url.String()}
	err = p.save()
	if err != nil {
		log.Fatalf("Error saving persist: %s", err)
	}

	err = srv.Install()
	if err != nil {
		log.Fatalf("Error installing service: %s", err)
	} else {
		log.Println("Service installed.")
	}
}

func uninstallCmdRun(cmd *cobra.Command, args []string) {
	err := srv.Uninstall()
	if err != nil {
		log.Fatalf("Error uninstalling service: %s", err)
	} else {
		log.Println("Service uninstalled.")
	}
}

func updateCmdRun(cmd *cobra.Command, args []string) {
	p := persist{}
	err := p.load()
	if err != nil {
		log.Fatalf("Error loading persist: %s", err)
	}
	status, err := srv.Status()
	if err != nil {
		log.Fatalf("Error getting service status: %s", err)
	}
	if status == service.StatusRunning {
		log.Println("Service is running, try to stop it...")
		err = srv.Stop()
		if err != nil {
			log.Fatalf("Error stopping service: %s", err)
		}
	}
	downloadBinary()
	downloadConfig(p.ConfigURL)
	err = srv.Start()
	if err != nil {
		log.Fatalf("Error starting service: %s", err)
	} else {
		log.Println("Service started.")
	}
}

func upgradeCmdRun(cmd *cobra.Command, args []string) {
	err := openBrowser("https://github.com/Zxilly/xxInstall/releases")
	if err != nil {
		log.Fatalf("Error opening browser: %s", err)
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch os := runtime.GOOS; os {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "xx",
		Short: "doing something",
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the xx",
		Run:   startCmdRun,
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the xx",
		Run:   stopCmdRun,
	}

	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the xx",
		Run:   restartCmdRun,
	}

	installCmd := &cobra.Command{
		Use:   "install [url]",
		Short: "Install the program with an config",
		Run:   installCmdRun,
		Args:  cobra.ExactArgs(1),
	}

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the xx",
		Run:   uninstallCmdRun,
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the xx",
		Run:   updateCmdRun,
	}

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the program itself",
		Run:   upgradeCmdRun,
	}

	rootCmd.AddCommand(startCmd, stopCmd, restartCmd, installCmd, uninstallCmd, updateCmd, upgradeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
