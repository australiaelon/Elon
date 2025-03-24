package libelon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sagernet/sing-box"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/deprecated"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	sJson "github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/service"
)

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

func GetVersionInfo() map[string]interface{} {
	return map[string]interface{}{
		"version": C.Version,
	}
}

func StartWithBase64Config(configBase64 string) (uint64, error) {
	jsonConfig, err := base64.StdEncoding.DecodeString(configBase64)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64 config: %w", err)
	}

	return StartWithJSONConfig(string(jsonConfig))
}

func StartWithJSONConfig(jsonConfig string) (uint64, error) {

	ctx, cancel := context.WithCancel(context.Background())
	ctx = service.ContextWithDefaultRegistry(ctx)
	ctx = service.ContextWith(ctx, deprecated.NewStderrManager(log.StdLogger()))
	ctx = box.Context(ctx, include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry())

	options, err := sJson.UnmarshalExtendedContext[option.Options](ctx, []byte(jsonConfig))
	if err != nil {
		cancel()
		return 0, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	instance, err := box.New(box.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		cancel()
		return 0, fmt.Errorf("failed to create instance: %w", err)
	}

	err = instance.Start()
	if err != nil {
		cancel()
		instance.Close()
		return 0, fmt.Errorf("failed to start instance: %w", err)
	}

	rbInstance := &Instance{
		box:    instance,
		config: jsonConfig,
		ctx:    ctx,
		cancel: cancel,
	}

	instancesLock.Lock()
	id := nextID
	nextID++
	instances[id] = rbInstance
	instancesLock.Unlock()

	return id, nil
}

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

	instance.cancel()

	err := instance.box.Close()
	if err != nil {
		return fmt.Errorf("failed to close instance: %w", err)
	}

	return nil
}

func GetInstance(instanceID uint64) (*Instance, bool) {
	instancesLock.Lock()
	defer instancesLock.Unlock()

	instance, exists := instances[instanceID]
	return instance, exists
}

func GetVersionBase64() (string, error) {
	versionInfo := GetVersionInfo()
	versionJSON, err := json.Marshal(versionInfo)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(versionJSON), nil
}

func StopAll() error {
	instancesLock.Lock()
	currentInstances := make(map[uint64]*Instance, len(instances))
	for id, instance := range instances {
		currentInstances[id] = instance
	}
	instancesLock.Unlock()

	var lastErr error
	for id := range currentInstances {
		err := Stop(id)
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}
