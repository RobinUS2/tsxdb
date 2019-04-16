package server

import (
	"../rpc/types"
	"log"
)

func init() {
	// init on module load
	registerEndpoint(NewReaderEndpoint())
}

type ReaderEndpoint struct {
}

func NewReaderEndpoint() *ReaderEndpoint {
	return &ReaderEndpoint{}
}

func (endpoint *ReaderEndpoint) Execute(args *types.ReadRequest, resp *types.ReadResponse) error {
	log.Printf("reader args %+v", args)
	// @todo implement read
	resp.Error = &types.RpcErrorNotImplemented
	return nil
}

func (endpoint *ReaderEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	return nil
}

func (endpoint *ReaderEndpoint) name() EndpointName {
	return EndpointName(types.EndpointReader)
}
