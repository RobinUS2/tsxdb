package server

import (
	"github.com/Route42/tsxdb/rpc/types"
	"github.com/Route42/tsxdb/server/backend"
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
	// auth
	if err := endpoint.server.validateSession(args.SessionTicket); err != nil {
		resp.Error = &types.RpcErrorAuthFailed
		return nil
	}

	// backend
	finalResults := make(map[uint64]map[uint64]float64)
	for _, query := range args.Queries {
		c := backend.ContextBackend{}
		c.Series = query.Id
		c.Namespace = query.Namespace
		backendInstance, err := endpoint.server.SelectBackend(c)
		if err != nil {
			resp.Error = &types.RpcErrorBackendStrategyNotFound
			return nil
		}

		// read
		readResult := backendInstance.Read(backend.ContextRead{Context: c.Context, From: query.From, To: query.To})
		if readResult.Error != nil {
			resp.Error = types.WrapErrorPointer(readResult.Error)
			return nil
		}
		// aggregation layer
		rollupResults := endpoint.server.rollupReader.Process(readResult)
		if rollupResults.Error != nil {
			resp.Error = types.WrapErrorPointer(rollupResults.Error)
			return nil
		}
		finalResults[query.Id] = rollupResults.Results
	}
	resp.Results = finalResults

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
