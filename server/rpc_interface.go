package server

import "sync"

var endpoints = make([]AbstractEndpoint, 0)
var endpointsMux sync.Mutex

type AbstractEndpoint interface {
	register(*EndpointOpts) error
	name() EndpointName
}

type EndpointName string

func (name EndpointName) String() string {
	return string(name)
}

type EndpointOpts struct {
	server *Instance
}
