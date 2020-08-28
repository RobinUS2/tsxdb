package server

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/server/backend"
	"strings"
	"sync"
	"sync/atomic"
)

func init() {
	// init on module load
	registerEndpoint(NewSeriesMetadataEndpoint())
}

type SeriesMetadataEndpoint struct {
	server    *Instance
	serverMux sync.RWMutex
}

func (endpoint *SeriesMetadataEndpoint) getServer() *Instance {
	endpoint.serverMux.RLock()
	s := endpoint.server
	endpoint.serverMux.RUnlock()
	return s
}

func NewSeriesMetadataEndpoint() *SeriesMetadataEndpoint {
	return &SeriesMetadataEndpoint{}
}

func (endpoint *SeriesMetadataEndpoint) Execute(args *types.SeriesMetadataRequest, resp *types.SeriesMetadataResponse) error {
	// deal with panics, else the whole RPC server could crash
	defer func() {
		if r := recover(); r != nil {
			resp.Error = types.WrapErrorPointer(fmt.Errorf("%s", r))
		}
	}()

	server := endpoint.getServer()

	// auth
	if err := server.validateSession(args.SessionTicket); err != nil {
		resp.Error = &types.RpcErrorAuthFailed
		return nil
	}

	// validate name
	if strings.Contains(args.SeriesCreateMetadata.Name, " ") {
		resp.Error = &types.RpcErrorSeriesNameWhitespace
		return nil
	}
	if len(args.SeriesCreateMetadata.Name) < 1 {
		resp.Error = &types.RpcErrorSeriesNameEmpty
		return nil
	}

	// metadata
	result := server.metaStore.CreateOrUpdateSeries(&backend.CreateSeries{
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

	// basic stats
	if resp.New {
		atomic.AddUint64(&server.numSeriesCreated, 1)
	} else {
		atomic.AddUint64(&server.numSeriesInitialised, 1)
	}

	return nil
}

func (endpoint *SeriesMetadataEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.serverMux.Lock()
	endpoint.server = opts.server
	endpoint.serverMux.Unlock()
	return nil
}

func (endpoint *SeriesMetadataEndpoint) name() EndpointName {
	return EndpointName(types.EndpointSeriesMetadata)
}
