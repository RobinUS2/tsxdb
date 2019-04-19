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
	// auth
	if err := endpoint.server.validateSession(args.SessionTicket); err != nil {
		resp.Error = &types.RpcErrorAuthFailed
		return nil
	}

	var numTimes int
	for _, batchItem := range args.Series {
		numTimes += len(batchItem.Times)
		numValues := len(batchItem.Values)
		if numTimes != numValues {
			resp.Error = &types.RpcErrorNumTimeValuePairsMisMatch
			return nil
		}

		// backend
		c := backend.ContextBackend{}
		c.Series = batchItem.Id
		c.Namespace = batchItem.Namespace
		backendInstance, err := endpoint.server.SelectBackend(c)
		if err != nil {
			resp.Error = &types.RpcErrorBackendStrategyNotFound
			return nil
		}

		// write
		writeContext := backend.ContextWrite{Context: c.Context}
		err = backendInstance.Write(writeContext, batchItem.Times, batchItem.Values)
		if err != nil {
			e := types.RpcError(err.Error())
			resp.Error = &e
			return nil
		}
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
