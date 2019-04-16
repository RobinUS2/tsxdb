package server

import "sync"

var endpoints = make([]AbstractEndpoint, 0)
var endpointsMux sync.Mutex

type AbstractEndpoint interface {
	register(*EndpointOpts) error
}

type EndpointOpts struct {
	server *Instance
}
