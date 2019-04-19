package server

import (
	"../rpc/types"
	"math/rand"
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
	// @todo real implementation
	resp.Id = rand.Uint64()

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
