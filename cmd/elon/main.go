package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

    "github.com/australiaelon/elon/libelon"
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

	fmt.Println("Features:")
	featuresMap := versionInfo["features"].(map[string]bool)

	features := []struct {
		name  string
		value bool
	}{
		{"IPv6 support", featuresMap["ipv6"]},
		{"gVisor support", featuresMap["gvisor"]},
		{"QUIC support", featuresMap["quic"]},
		{"Wireguard support", featuresMap["wireguard"]},
		{"ECH support", featuresMap["ech"]},
		{"UTLS support", featuresMap["utls"]},
		{"Clash API support", featuresMap["clash_api"]},
		{"V2Ray API support", featuresMap["v2ray_api"]},
		{"ShadowsocksR support", featuresMap["ssr"]},
		{"DHCP support", featuresMap["dhcp"]},
		{"Low Memory support", featuresMap["low_memory"]},
		{"Connection tracking support", featuresMap["conntrack"]},
		{"System service support", featuresMap["system_service"]},
	}

	for _, feature := range features {
		fmt.Printf("  %s: %t\n", feature.name, feature.value)
	}

	if featuresMap["cgo_enabled"] {
		fmt.Println("CGO: enabled")
	} else {
		fmt.Println("CGO: disabled")
	}
}
