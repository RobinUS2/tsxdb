package client

import (
	"../rpc/types"
	"github.com/pkg/errors"
	"sync/atomic"
)

// init series metadata on the server side (only transmits names, tags, etc. once instead of each time)

func (series *Series) Init(conn *ManagedConnection) (err error) {
	// fast path
	if atomic.LoadUint64(&series.id) > 0 {
		// already initialised
		return nil
	}

	// verify connection
	if conn == nil {
		return errors.New("missing connection")
	}

	// max 1 init at a time
	series.initMux.Lock()
	defer series.initMux.Unlock()

	// check again already sent? (could be done during waiting of the lock)
	if atomic.LoadUint64(&series.id) > 0 {
		// already initialised
		return nil
	}

	// request
	request := types.SeriesMetadataRequest{
		Namespace:     series.namespace,
		Tags:          series.tags,
		Name:          series.name,
		SessionTicket: conn.getSessionTicket(),
	}

	// execute
	var response *types.SeriesMetadataResponse
	if err := conn.client.Call(types.EndpointSeriesMetadata.String()+"."+types.MethodName, request, &response); err != nil {
		return err
	}
	if response.Error != nil {
		return response.Error.Error()
	}

	// store id
	atomic.StoreUint64(&series.id, response.Id)

	return nil
}
