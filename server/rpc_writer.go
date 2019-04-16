package server

import (
	"../rpc/types"
	"log"
)

type WriterEndpoint struct {
}

func (endpoint *WriterEndpoint) Write(args *types.WriteRequest, resp *types.WriteResponse) error {
	log.Printf("args %+v", args)
	resp.Num = len(args.Times)
	resp.Error = &types.RpcErrorNotImplemented
	// @todo implement
	return nil
}

func (endpoint *WriterEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.Register(endpoint); err != nil {
		return err
	}
	return nil
}

func init() {
	endpoint := &WriterEndpoint{}
	endpointsMux.Lock()
	endpoints = append(endpoints, endpoint)
	endpointsMux.Unlock()
}
