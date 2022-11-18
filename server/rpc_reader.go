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
	registerEndpoint(NewReaderEndpoint())
}

type ReaderEndpoint struct {
	server    *Instance
	serverMux sync.RWMutex
}

func (endpoint *ReaderEndpoint) getServer() *Instance {
	endpoint.serverMux.RLock()
	s := endpoint.server
	endpoint.serverMux.RUnlock()
	return s
}

func NewReaderEndpoint() *ReaderEndpoint {
	return &ReaderEndpoint{}
}

func (endpoint *ReaderEndpoint) Execute(args *types.ReadRequest, resp *types.ReadResponse) error {
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

	// backend
	finalResults := make(map[uint64]map[uint64]float64)
	for _, query := range args.Queries {
		c := backend.ContextBackend{}
		c.Series = query.Id
		c.Namespace = query.Namespace
		backendInstance, err := server.SelectBackend(c)
		if err != nil {
			resp.Error = &types.RpcErrorBackendStrategyNotFound
			return nil
		}

		// read
		readResult := backendInstance.Read(backend.ContextRead{Context: c.Context, From: query.From, To: query.To})
		if readResult.Error != nil && !strings.Contains(readResult.Error.Error(), types.RpcErrorNoDataFound.String()) {
			// return all errors, except if no data found, since we can query 1-N series, 1 series no data is not a fatal error
			resp.Error = types.WrapErrorPointer(readResult.Error)
			return nil
		}
		// aggregation layer
		rollupResults := server.rollupReader.Process(readResult)
		if rollupResults.Error != nil {
			resp.Error = types.WrapErrorPointer(rollupResults.Error)
			return nil
		}
		finalResults[query.Id] = rollupResults.Results
	}
	resp.Results = finalResults

	// stats
	atomic.AddUint64(&server.numReads, 1)

	return nil
}

func (endpoint *ReaderEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.serverMux.Lock()
	endpoint.server = opts.server
	endpoint.serverMux.Unlock()
	return nil
}

func (endpoint *ReaderEndpoint) name() EndpointName {
	return EndpointName(types.EndpointReader)
}
