package server

import (
	"log"
	"net/rpc"
)

type Instance struct {
	opts *Opts
	rpc  *rpc.Server
}

func New(opts *Opts) *Instance {
	return &Instance{
		opts: opts,
		rpc:  rpc.NewServer(),
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
	return nil
}
