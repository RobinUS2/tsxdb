package server

import (
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/server/backend"
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

	var numTimesTotal int
	for _, batchItem := range args.Series {
		numTimes := len(batchItem.Times)
		numTimesTotal += numTimes
		numValues := len(batchItem.Values)

		// basic validation
		if numTimes < 1 {
			resp.Error = &types.RpcErrorNoValues
			return nil
		}
		if numTimes != numValues {
			resp.Error = &types.RpcErrorNumTimeValuePairsMisMatch
			return nil
		}
		if batchItem.Id < 1 {
			resp.Error = &types.RpcErrorMissingSeriesId
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
		writeContext := backend.ContextWrite(c)
		err = backendInstance.Write(writeContext, batchItem.Times, batchItem.Values)
		if err != nil {
			e := types.RpcError(err.Error())
			resp.Error = &e
			return nil
		}
	}
	resp.Num = numTimesTotal

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
