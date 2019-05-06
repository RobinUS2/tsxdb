package server

import (
	"github.com/Route42/tsxdb/rpc/types"
	"github.com/Route42/tsxdb/server/backend"
)

func init() {
	// init on module load
	registerEndpoint(NewSeriesMetadataEndpoint())
}

type SeriesMetadataEndpoint struct {
	server *Instance
}

func NewSeriesMetadataEndpoint() *SeriesMetadataEndpoint {
	return &SeriesMetadataEndpoint{}
}

func (endpoint *SeriesMetadataEndpoint) Execute(args *types.SeriesMetadataRequest, resp *types.SeriesMetadataResponse) error {
	// auth
	if err := endpoint.server.validateSession(args.SessionTicket); err != nil {
		resp.Error = &types.RpcErrorAuthFailed
		return nil
	}

	// metadata
	result := endpoint.server.metaStore.CreateOrUpdateSeries(&backend.CreateSeries{
		Series: map[types.SeriesCreateIdentifier]types.SeriesCreateMetadata{
			args.SeriesCreateIdentifier: args.SeriesCreateMetadata,
		},
	})
	thisResult := result.Results[args.SeriesCreateIdentifier] // only support one for now
	// for some reason assigning thisResult to resp is not working, probably since the reference is part of the RPC pipe
	resp.New = thisResult.New
	resp.Id = thisResult.Id
	resp.Error = thisResult.Error
	resp.SeriesCreateIdentifier = thisResult.SeriesCreateIdentifier

	return nil
}

func (endpoint *SeriesMetadataEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.server = opts.server
	return nil
}

func (endpoint *SeriesMetadataEndpoint) name() EndpointName {
	return EndpointName(types.EndpointSeriesMetadata)
}
