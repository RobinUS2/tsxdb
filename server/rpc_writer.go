package server

import (
	"../rpc/types"
	"./backend"
	"log"
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
	log.Printf("writer args %+v", args)
	numTimes := len(args.Times)
	numValues := len(args.Values)
	if numTimes != numValues {
		resp.Error = &types.RpcErrorNumTimeValuePairsMisMatch
		return nil
	}

	// select
	selectedStrategy, err := endpoint.server.backendSelector.SelectStrategy(backend.Context{})
	if err != nil {
		resp.Error = &types.RpcErrorBackendStrategyNotFound
		return nil
	}
	b := selectedStrategy.GetBackend()

	// write
	for idx, ts := range args.Times {
		val := args.Values[idx]
		err := b.Write(args.SeriesIdentifier.Namespace, args.SeriesIdentifier.Id, ts, val)
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
