package server

import (
	"./backend"
	"./rollup"
	"log"
	"net/rpc"
)

type Instance struct {
	opts            *Opts
	rpc             *rpc.Server
	backendSelector *backend.Selector
	rollupReader    *rollup.Reader
}

func New(opts *Opts) *Instance {
	return &Instance{
		opts:         opts,
		rpc:          rpc.NewServer(),
		rollupReader: rollup.NewReader(),
	}
}

func (instance *Instance) Init() error {
	// register all endpoints
	endpointOpts := &EndpointOpts{server: instance}
	for _, endpoint := range endpoints {
		if err := endpoint.register(endpointOpts); err != nil {
			return err
		}
	}
	log.Printf("%d endpoints registered", len(endpoints))

	// testing backend strategy in memory
	// @todo from config
	instance.backendSelector = backend.NewSelector()
	myStrategy := backend.NewSimpleStrategy(backend.NewMemoryBackend())
	if err := instance.backendSelector.AddStrategy(myStrategy); err != nil {
		return err
	}

	return nil
}

func (instance *Instance) Start() error {
	if err := instance.StartListening(); err != nil {
		return err
	}
	return nil
}
