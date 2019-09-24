package server

import (
	"errors"
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
)

func init() {
	// init on module load
	registerEndpoint(NewNoOpEndpoint())
}

type NoOpEndpoint struct {
	server *Instance
}

func NewNoOpEndpoint() *NoOpEndpoint {
	return &NoOpEndpoint{}
}

func (endpoint *NoOpEndpoint) Execute(args *types.ReadRequest, resp *types.ReadResponse) error {
	// deal with panics, else the whole RPC server could crash
	defer func() {
		if r := recover(); r != nil {
			resp.Error = types.WrapErrorPointer(errors.New(fmt.Sprintf("%s", r)))
		}
	}()

	// auth
	if err := endpoint.server.validateSession(args.SessionTicket); err != nil {
		resp.Error = &types.RpcErrorAuthFailed
		return nil
	}
	return nil
}

func (endpoint *NoOpEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.server = opts.server
	return nil
}

func (endpoint *NoOpEndpoint) name() EndpointName {
	return EndpointName(types.EndpointNoOp)
}
