package main

import (
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"log"
	"net/url"
	"os/exec"
	"runtime"
)

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
	t, err := url.Parse(u)
	if err != nil {
		log.Fatalf("Error parsing url: %s", err)
	}

	ondemand := cmd.Flag("ondemand").Value.String() == "true"

	downloadConfig(t.String())
	if !isExecutableExist("sing-box") || !ondemand {
		downloadBinary()
	}

	p := persist{ConfigURL: t.String()}
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

func serviceCmdRun(cmd *cobra.Command, args []string) {
	err := srv.Run()
	if err != nil {
		log.Fatalf("Error running service: %s", err)
	}
	return
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

var rootCmd = &cobra.Command{
	Use:   "xx",
	Short: "do something",
}

func init() {
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
	installCmd.Flags().Bool("ondemand", true, "Only download the binary if necessary")

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

	serviceCmd := &cobra.Command{
		Use:    "service",
		Run:    serviceCmdRun,
		Hidden: true,
	}

	rootCmd.AddCommand(startCmd, stopCmd, restartCmd, installCmd, uninstallCmd, updateCmd, upgradeCmd, serviceCmd)
}