package server

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/server/backend"
	"sync"
	"sync/atomic"
)

func init() {
	// init on module load
	registerEndpoint(NewWriterEndpoint())
}

type WriterEndpoint struct {
	server    *Instance
	serverMux sync.RWMutex
}

func (endpoint *WriterEndpoint) getServer() *Instance {
	endpoint.serverMux.RLock()
	s := endpoint.server
	endpoint.serverMux.RUnlock()
	return s
}

func NewWriterEndpoint() *WriterEndpoint {
	return &WriterEndpoint{}
}

func (endpoint *WriterEndpoint) Execute(args *types.WriteRequest, resp *types.WriteResponse) error {
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

	// request ID to track this specific request
	requestId := backend.NewRequestId()

	// track backend instances for final flushes
	backendInstances := make(map[backend.IAbstractBackend]bool)

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
		c.RequestId = requestId
		backendInstance, err := server.SelectBackend(c)
		if err != nil {
			resp.Error = &types.RpcErrorBackendStrategyNotFound
			return nil
		}
		backendInstances[backendInstance] = true

		// write
		writeContext := backend.ContextWrite(c)
		err = backendInstance.Write(writeContext, batchItem.Times, batchItem.Values)
		if err != nil {
			e := types.RpcError(err.Error())
			resp.Error = &e
			return nil
		}
	}

	// flush backends
	for backendInstance := range backendInstances {
		if err := backendInstance.FlushPendingWrites(requestId); err != nil {
			e := types.RpcError(err.Error())
			resp.Error = &e
			return nil
		}
	}

	resp.Num = numTimesTotal

	// basic stats
	atomic.AddUint64(&server.numValuesWritten, uint64(resp.Num))

	return nil
}

func (endpoint *WriterEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.serverMux.Lock()
	endpoint.server = opts.server
	endpoint.serverMux.Unlock()
	return nil
}

func (endpoint *WriterEndpoint) name() EndpointName {
	return EndpointName(types.EndpointWriter)
}
