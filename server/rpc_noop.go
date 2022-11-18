package server

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"sync"
)

func init() {
	// init on module load
	registerEndpoint(NewNoOpEndpoint())
}

type NoOpEndpoint struct {
	server    *Instance
	serverMux sync.RWMutex
}

func (endpoint *NoOpEndpoint) getServer() *Instance {
	endpoint.serverMux.RLock()
	s := endpoint.server
	endpoint.serverMux.RUnlock()
	return s
}

func NewNoOpEndpoint() *NoOpEndpoint {
	return &NoOpEndpoint{}
}

func (endpoint *NoOpEndpoint) Execute(args *types.ReadRequest, resp *types.ReadResponse) error {
	// deal with panics, else the whole RPC server could crash
	defer func() {
		if r := recover(); r != nil {
			resp.Error = types.WrapErrorPointer(fmt.Errorf("%s", r))
		}
	}()

	// auth
	server := endpoint.getServer()
	if err := server.validateSession(args.SessionTicket); err != nil {
		resp.Error = &types.RpcErrorAuthFailed
		return nil
	}
	return nil
}

func (endpoint *NoOpEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.serverMux.Lock()
	endpoint.server = opts.server
	endpoint.serverMux.Unlock()
	return nil
}

func (endpoint *NoOpEndpoint) name() EndpointName {
	return EndpointName(types.EndpointNoOp)
}
