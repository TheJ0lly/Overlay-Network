package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var ovnetoPathDir string

func startNodeMgmt(cmd *cobra.Command, args []string) {
	nodemgmtInfoFile := fmt.Sprintf("%s/nodemgmt.info", ovnetoPathDir)
	if _, err := os.Stat(nodemgmtInfoFile); err == nil {
		fmt.Printf("nodemgmt service already running\n")
		return
	}

	port := "8080"
	if len(args) > 0 {
		port = args[0]
	}

	// We start the nodemgmt service from which we will control the nodes
	process := exec.Command(fmt.Sprintf("%s/nodemgmt", ovnetoPathDir), "-port", port)

	process.Stdout = os.Stdout
	process.Stdin = os.Stdin

	process.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	err := process.Start()
	if err != nil {
		fmt.Printf("error while starting nodemgmt service: %s\n", err)
		return
	}

	f, err := os.Create(nodemgmtInfoFile)
	if err != nil {
		fmt.Printf("error while saving nodemgmt service info: %s\n", err)
		return
	}

	if _, err = fmt.Fprintf(f, "%d\n%s", process.Process.Pid, port); err != nil {
		fmt.Printf("error while saving nodemgmt service info: %s\n", err)
		return
	}
}

func stopNodeMgmt(cmd *cobra.Command, args []string) {
	nodemgmtInfoFile := fmt.Sprintf("%s/nodemgmt.info", ovnetoPathDir)
	if _, err := os.Stat(nodemgmtInfoFile); err != nil {
		fmt.Printf("nodemgmt service is not running\n")
		return
	}

	b, err := os.ReadFile(nodemgmtInfoFile)
	if err != nil {
		fmt.Printf("error while reading nodemgmt service info: %s\n", err)
		return
	}

	serviceInfo := strings.Split(string(b), "\n")
	pid, err := strconv.Atoi(serviceInfo[0])
	if err != nil {
		fmt.Printf("error while reading nodemgmt service info: %s\n", err)
		return
	}

	if err = os.Remove(nodemgmtInfoFile); err != nil {
		fmt.Printf("error while deleting nodemgmt service info: %s - please delete it manually, if possible\n", err)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("error while finding nodemgmt service: %s\n", err)
		return
	}

	err = p.Kill()
	if err != nil {
		fmt.Printf("error while killing nodemgmt service: %s\n", err)
		return
	}
}

func main() {
	ovnetoPath, err := os.Executable()
	if err != nil {
		fmt.Printf("could not get the ovneto absolute path: %s - cannot use any related feature\n", err)
		return
	}

	ovnetoPathDir, err = filepath.Abs(ovnetoPath)
	if err != nil {
		fmt.Printf("could not get the ovneto absolute path: %s - cannot use any related feature\n", err)
		return
	}

	ovnetoPathDir = filepath.Dir(ovnetoPath)

	rootCmd := &cobra.Command{Use: "ovneto [start | stop]"}
	startCmd := &cobra.Command{
		Use:   "start [port=(default 8080)]",
		Short: "starts the nodemgmt service",
		Args:  cobra.MaximumNArgs(1),
		Run:   startNodeMgmt,
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "stops the nodemgmt service",
		Run:   stopNodeMgmt,
	}

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	err = rootCmd.Execute()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
