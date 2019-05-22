package client

import (
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/tools"
	"github.com/pkg/errors"
	"sync/atomic"
)

// init series metadata on the server side (only transmits names, tags, etc. once instead of each time)

func (series *Series) Init(conn *ManagedConnection) (id uint64, err error) {
	// fast path
	id = atomic.LoadUint64(&series.id)
	if id > 0 {
		// already initialised
		return id, nil
	}

	// verify connection
	if conn == nil {
		return 0, errors.New("missing connection")
	}

	// max 1 init at a time
	series.initMux.Lock()
	defer series.initMux.Unlock()

	// check again already sent? (could be done during waiting of the lock)
	if atomic.LoadUint64(&series.id) > 0 {
		// already initialised
		return 0, nil
	}

	// request
	request := types.SeriesMetadataRequest{
		SeriesCreateMetadata: types.SeriesCreateMetadata{
			SeriesMetadata: types.SeriesMetadata{
				Namespace: series.namespace,
				Tags:      series.tags,
				Name:      series.name,
				Ttl:       series.ttl,
			},
			SeriesCreateIdentifier: types.SeriesCreateIdentifier(tools.RandomInsecureIdentifier()),
		},
		SessionTicket: conn.getSessionTicket(),
	}

	// execute
	var response *types.SeriesMetadataResponse
	if err := conn.client.Call(types.EndpointSeriesMetadata.String()+"."+types.MethodName, request, &response); err != nil {
		return 0, err
	}
	if response.Error != nil {
		return 0, response.Error.Error()
	}

	// store id
	atomic.StoreUint64(&series.id, response.Id)

	return series.id, nil
}
