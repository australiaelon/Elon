package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

    "github.com/australiaelon/Elon/libelon"
)

var (
	configBase64 string
	versionFlag  bool
)

func init() {
	flag.StringVar(&configBase64, "c", "", "Base64 encoded configuration")
	flag.BoolVar(&versionFlag, "version", false, "Show version information")
}

func main() {
	flag.Parse()

	if versionFlag {
		printVersion()
		return
	}

	if configBase64 == "" {
		fmt.Println("Error: -c with base64 configuration is required")
		flag.Usage()
		os.Exit(1)
	}

	instanceID, err := libelon.StartWithBase64Config(configBase64)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("elon instance started with ID: %d\n", instanceID)

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	<-osSignals

	fmt.Println("Shutting down...")

	err = libelon.Stop(instanceID)
	if err != nil {
		fmt.Printf("Error stopping instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Gracefully stopped")
}

func printVersion() {
	versionInfo := libelon.GetVersionInfo()
	fmt.Printf("elon version: %s\n\n", versionInfo["version"])
}
