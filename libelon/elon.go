package libelon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sagernet/sing-box"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

// Instance represents a running elon instance
type Instance struct {
	box    *box.Box
	config string
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	instances     = make(map[uint64]*Instance)
	instancesLock sync.Mutex
	nextID        uint64 = 1
)

// GetVersionInfo returns version and feature information
func GetVersionInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":  C.Version,
	}
}

// StartWithBase64Config starts a new instance with base64-encoded configuration
func StartWithBase64Config(configBase64 string) (uint64, error) {
	// Decode the base64 configuration
	jsonConfig, err := base64.StdEncoding.DecodeString(configBase64)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64 config: %w", err)
	}

	// Start with the decoded JSON config
	return StartWithJSONConfig(string(jsonConfig))
}

// StartWithJSONConfig starts a new instance with JSON configuration
func StartWithJSONConfig(jsonConfig string) (uint64, error) {
	// Parse the JSON configuration
	var options option.Options
	err := json.Unmarshal([]byte(jsonConfig), &options)
	if err != nil {
		return 0, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create the instance
	instance, err := box.New(box.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		cancel()
		return 0, fmt.Errorf("failed to create instance: %w", err)
	}

	// Start the instance
	err = instance.Start()
	if err != nil {
		cancel()
		instance.Close()
		return 0, fmt.Errorf("failed to start instance: %w", err)
	}

	// Create and store our managed instance
	rbInstance := &Instance{
		box:    instance,
		config: jsonConfig,
		ctx:    ctx,
		cancel: cancel,
	}

	// Store the instance with a unique ID
	instancesLock.Lock()
	id := nextID
	nextID++
	instances[id] = rbInstance
	instancesLock.Unlock()

	return id, nil
}

// Stop stops a running instance by ID
func Stop(instanceID uint64) error {
	instancesLock.Lock()
	instance, exists := instances[instanceID]
	if exists {
		delete(instances, instanceID)
	}
	instancesLock.Unlock()

	if !exists {
		return fmt.Errorf("instance not found")
	}

	// Cancel the context
	instance.cancel()

	// Close the instance
	err := instance.box.Close()
	if err != nil {
		return fmt.Errorf("failed to close instance: %w", err)
	}

	return nil
}

// GetInstance retrieves an instance by ID
func GetInstance(instanceID uint64) (*Instance, bool) {
	instancesLock.Lock()
	defer instancesLock.Unlock()
	
	instance, exists := instances[instanceID]
	return instance, exists
}

// GetVersionBase64 returns version information as base64-encoded JSON
func GetVersionBase64() (string, error) {
	versionInfo := GetVersionInfo()
	versionJSON, err := json.Marshal(versionInfo)
	if err != nil {
		return "", err
	}
	
	return base64.StdEncoding.EncodeToString(versionJSON), nil
}