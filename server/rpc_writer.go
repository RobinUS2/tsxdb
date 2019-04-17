package server

import (
	"../rpc/types"
	"./backend"
)

func init() {
	// init on module load
	registerEndpoint(NewWriterEndpoint())
}

type WriterEndpoint struct {
	server *Instance
}

func NewWriterEndpoint() *WriterEndpoint {
	return &WriterEndpoint{}
}

func (endpoint *WriterEndpoint) Execute(args *types.WriteRequest, resp *types.WriteResponse) error {
	numTimes := len(args.Times)
	numValues := len(args.Values)
	if numTimes != numValues {
		resp.Error = &types.RpcErrorNumTimeValuePairsMisMatch
		return nil
	}

	// backend
	c := backend.ContextBackend{}
	c.Series = args.SeriesIdentifier.Id
	// @todo namespace
	backendInstance, err := endpoint.server.SelectBackend(c)
	if err != nil {
		resp.Error = &types.RpcErrorBackendStrategyNotFound
		return nil
	}

	// write
	writeContext := backend.ContextWrite{Context: c.Context}
	err = backendInstance.Write(writeContext, args.Times, args.Values)
	if err != nil {
		e := types.RpcError(err.Error())
		resp.Error = &e
		return nil
	}
	resp.Num = numTimes

	return nil
}

func (endpoint *WriterEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.server = opts.server
	return nil
}

func (endpoint *WriterEndpoint) name() EndpointName {
	return EndpointName(types.EndpointWriter)
}
