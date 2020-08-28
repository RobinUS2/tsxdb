package server

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/server/backend"
	"github.com/pkg/errors"
	"log"
	"strings"
	"time"
)

// initialise all server stuff without actually listening
func (instance *Instance) Init() error {
	// register all endpoints
	endpointOpts := &EndpointOpts{server: instance}
	for _, endpoint := range endpoints {
		if err := endpoint.register(endpointOpts); err != nil {
			return err
		}
	}

	// default backend?
	if len(instance.opts.Backends) < 1 {
		// default backend memory
		log.Printf("WARN no backends defined, creating default non-persistent embedded memory backend")
		instance.opts.Backends = []BackendOpts{
			{
				Identifier: backend.DefaultIdentifier,
				Type:       backend.MemoryType.String(),
				Metadata:   true,
			},
		}
	}

	// create backends
	backends := make([]backend.IAbstractBackend, 0)
	for _, backendOpt := range instance.opts.Backends {
		b := backend.InstanceFactory(backendOpt.Type, backendOpt.Options)
		if b == nil {
			panic(fmt.Sprintf("failed to construct backend %+v", backendOpt))
		}
		backends = append(backends, b)
	}
	if len(backends) > 1 {
		return errors.New("no more than 1 backend supported for now")
	}

	// backend strategy
	if len(strings.TrimSpace(instance.opts.BackendStrategy.Type)) < 1 {
		instance.opts.BackendStrategy.Type = backend.SimpleStrategyType.String()
	}
	myStrategy := backend.StrategyInstanceFactory(instance.opts.BackendStrategy.Type, instance.opts.BackendStrategy.Options)
	myStrategy.SetBackends(backends)

	// backend selector
	instance.backendSelector = backend.NewSelector()
	if err := instance.backendSelector.AddStrategy(myStrategy); err != nil {
		return err
	}

	// must have auth
	if len(strings.TrimSpace(instance.opts.AuthToken)) < 1 {
		return errors.New("missing mandatory auth token option")
	}

	// metadata
	var metadataBackend backend.AbstractBackendWithMetadata
	for i, backendInstance := range backends {
		backendOpts := instance.opts.Backends[i]
		if !backendOpts.Metadata {
			continue
		}
		if typed, ok := backendInstance.(backend.AbstractBackendWithMetadata); ok {
			metadataBackend = typed
		} else {
			panic(fmt.Sprintf("backend %+v claims incorrectly to be able to store metadata", backendOpts))
		}
	}
	if metadataBackend == nil {
		return errors.New("no metadata backend defined")
	}
	instance.metaStore = backend.NewMetadata(metadataBackend)

	// link backends back to the system
	for _, backendInstance := range backends {
		backendInstance.SetReverseApi(instance.metaStore)
	}

	// init backends
	for _, backendInstance := range backends {
		if err := backendInstance.Init(); err != nil {
			return err
		}
	}

	// stats ticker
	instance.statsTicker = time.NewTicker(60 * time.Second)
	go func() {
		for range instance.statsTicker.C {
			log.Printf("stats %+v", instance.Statistics())
		}
	}()

	// session ticker
	instance.sessionTicker = time.NewTicker(300 * time.Millisecond)
	go func() {
		for range instance.sessionTicker.C {
			instance.sessionExpire()
		}
	}()

	// connection ticker
	instance.connectionTicker = time.NewTicker(300 * time.Millisecond)
	go func() {
		for range instance.connectionTicker.C {
			instance.connectionExpire()
		}
	}()

	return nil
}
