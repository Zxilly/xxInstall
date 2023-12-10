package main

import (
	"bufio"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"time"
)

func startCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
	err := srv.Start()
	if err != nil {
		log.Fatalf("Error starting service: %s", err)
	} else {
		log.Println("Service started.")
	}
}

func stopCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
	err := srv.Stop()
	if err != nil {
		log.Fatalf("Error stopping service: %s", err)
	} else {
		log.Println("Service stopped.")
	}
}

func restartCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
	err := srv.Restart()
	if err != nil {
		log.Fatalf("Error restarting service: %s", err)
	} else {
		log.Println("Service restarted.")
	}
}

func statusToString(status service.Status) string {
	switch status {
	case service.StatusRunning:
		return "Running"
	case service.StatusStopped:
		return "Stopped"
	case service.StatusUnknown:
		return "Unknown"
	default:
		return "Unexpected"
	}
}

func statusCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
	status, err := srv.Status()
	if err != nil {
		log.Fatalf("Error getting service status: %s", err)
	}
	log.Printf("Service status: %s", statusToString(status))
}

func syncCmdRun(command *cobra.Command, args []string) {
	requireRoot()

	state, err := srv.Status()
	if err != nil {
		log.Fatalf("Error getting service status: %s", err)
	}

	shouldRestart := false
	if state == service.StatusRunning {
		err = srv.Stop()
		if err != nil {
			log.Fatalf("Error stopping service: %s", err)
		}
		log.Println("Service stopped.")
		shouldRestart = true
	}

	p := persist{}
	err = p.load()
	if err != nil {
		log.Fatalf("Error loading persist: %s", err)
	}
	if p.ConfigURL == "" {
		log.Fatalf("No config url found")
	}
	downloadConfig(p.ConfigURL)

	if shouldRestart {
		err = srv.Start()
		if err != nil {
			log.Fatalf("Error starting service: %s", err)
		} else {
			log.Println("Service started.")
		}
	}
}

func installCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
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
	system := cmd.Flag("system").Value.String() == "true"

	downloadConfig(t.String())

	if system {
		// find sing-box
		_, err := exec.LookPath("sing-box")
		if err != nil {
			log.Fatalf("Error finding sing-box: %s", err)
		}
	} else {
		if !isExecutableExist("sing-box") || !ondemand {
			prerelease := cmd.Flag("prerelease").Value.String() == "true"
			version := cmd.Flag("version").Value.String()
			mirror := cmd.Flag("mirror").Value.String() == "true"
			downloadBinary(prerelease, version, mirror)
		}
	}

	p := persist{
		ConfigURL:           t.String(),
		PreferSystemInstall: system,
	}
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

func logCmdRun(cmd *cobra.Command, args []string) {
	// check LogFile
	_, err := os.Stat(LogFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Log file %s not exist", LogFile)
		} else {
			log.Fatalf("Error checking log file: %s", err)
		}
	}

	// read the last 10 lines
	file, err := os.Open(LogFile)
	if err != nil {
		log.Fatalf("Error opening log file: %s", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, line)
		if len(lines) > 20 {
			lines = lines[1:]
		}
	}
	for _, line := range lines {
		log.Print(line)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Error getting log file info: %s", err)
	}

	size := fileInfo.Size()

	for {
		_, err := file.Seek(size, io.SeekStart)
		if err != nil {
			log.Fatalf("Error seeking log file: %s", err)
		}

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			log.Print(line)
		}

		fileInfo, err := file.Stat()
		if err != nil {
			log.Fatalf("Error getting log file info: %s", err)
		}
		size = fileInfo.Size()

		time.Sleep(1 * time.Second)
	}
}

func uninstallCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
	err := srv.Uninstall()
	if err != nil {
		log.Fatalf("Error uninstalling service: %s", err)
	} else {
		log.Println("Service uninstalled.")
	}
}

func updateCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
	p := persist{}
	err := p.load()
	if err != nil {
		log.Fatalf("Error loading persist: %s", err)
	}

	shouldRestart := false
	status, err := srv.Status()
	if err != nil {
		log.Fatalf("Error getting service status: %s", err)
	}
	if status == service.StatusRunning {
		shouldRestart = true
		log.Println("Service is running, try to stop it...")
		err = srv.Stop()
		if err != nil {
			log.Fatalf("Error stopping service: %s", err)
		}
	}

	prerelease := cmd.Flag("prerelease").Value.String() == "true"
	version := cmd.Flag("version").Value.String()
	mirror := cmd.Flag("mirror").Value.String() == "true"
	downloadBinary(prerelease, version, mirror)

	if shouldRestart {
		log.Println("Try to restart service")
		err = srv.Start()
		if err != nil {
			log.Fatalf("Error starting service: %s", err)
		} else {
			log.Println("Service started.")
		}
	}
}

func upgradeCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()

	status, err := srv.Status()
	if err != nil {
		log.Fatalf("Error getting service status: %s", err)
	}

	shouldRestart := false
	if status == service.StatusRunning {
		shouldRestart = true
		log.Println("Service is running, try to stop it...")
		err = srv.Stop()
		if err != nil {
			log.Fatalf("Error stopping service: %s", err)
		}
	}

	mirror := cmd.Flag("mirror").Value.String() == "true"
	applySelfUpdate(mirror)

	if shouldRestart {
		log.Println("Try to restart service")
		err = srv.Start()
		if err != nil {
			log.Fatalf("Error starting service: %s", err)
		} else {
			log.Println("Service started.")
		}
	}
}

func serviceCmdRun(cmd *cobra.Command, args []string) {
	requireRoot()
	err := srv.Run()
	if err != nil {
		log.Fatalf("Error running service: %s", err)
	}
	return
}

var rootCmd = &cobra.Command{
	Use:   "xx",
	Short: "do something",
}

func applyDownloadFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("prerelease", "p", true, "Allow to install prerelease version")
	cmd.Flags().StringP("version", "v", "", "Specify the version to install")
	applyMirrorFlag(cmd)
}

func applyMirrorFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("mirror", "m", false, "Use mirror to download")
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
	installCmd.Flags().Bool("system", false, "Prefer to find binary in the system path")
	applyDownloadFlag(installCmd)

	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Show the log",
		Run:   logCmdRun,
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
	applyDownloadFlag(updateCmd)

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the program itself",
		Run:   upgradeCmdRun,
	}
	applyMirrorFlag(upgradeCmd)

	serviceCmd := &cobra.Command{
		Use:    "service",
		Run:    serviceCmdRun,
		Hidden: true,
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get the status of the xx",
		Run:   statusCmdRun,
	}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync the config",
		Run:   syncCmdRun,
	}

	rootCmd.AddCommand(
		startCmd,
		stopCmd,
		restartCmd,
		installCmd,
		logCmd,
		uninstallCmd,
		updateCmd,
		upgradeCmd,
		serviceCmd,
		statusCmd,
		syncCmd)
}
