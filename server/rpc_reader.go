package server

import (
	"../rpc/types"
	"./backend"
)

func init() {
	// init on module load
	registerEndpoint(NewReaderEndpoint())
}

type ReaderEndpoint struct {
	server *Instance
}

func NewReaderEndpoint() *ReaderEndpoint {
	return &ReaderEndpoint{}
}

func (endpoint *ReaderEndpoint) Execute(args *types.ReadRequest, resp *types.ReadResponse) error {
	// backend
	c := backend.ContextBackend{}
	c.Series = args.SeriesIdentifier.Id
	// @todo namespace
	backendInstance, err := endpoint.server.SelectBackend(c)
	if err != nil {
		resp.Error = &types.RpcErrorBackendStrategyNotFound
		return nil
	}

	// read
	readResult := backendInstance.Read(backend.ContextRead{Context: c.Context, From: args.From, To: args.To})
	if readResult.Error != nil {
		resp.Error = types.WrapErrorPointer(readResult.Error)
		return nil
	}
	// aggregation layer
	rollupResults := endpoint.server.rollupReader.Process(readResult)
	resp.Results = rollupResults.Results

	return nil
}

func (endpoint *ReaderEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.server = opts.server
	return nil
}

func (endpoint *ReaderEndpoint) name() EndpointName {
	return EndpointName(types.EndpointReader)
}
