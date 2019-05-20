package client

import "github.com/RobinUS2/tsxdb/rpc/types"

type BatchWriter struct {
	client *Instance
	items  []BatchItem
}

func (batch *BatchWriter) ToWriteRequest(conn *ManagedConnection) (request types.WriteRequest, err error) {
	seriesTimestamps := make(map[uint64][]uint64) // key of outer slice is the series id
	seriesValues := make(map[uint64][]float64)    // key of outer slice is the series id
	seriesNamespace := make(map[uint64]int)       // key of outer slice is the series id
	for _, item := range batch.items {
		var seriesId uint64
		if seriesId, err = item.series.Init(conn); err != nil {
			return
		}
		if seriesTimestamps[seriesId] == nil {
			seriesTimestamps[seriesId] = make([]uint64, 0)
			seriesValues[seriesId] = make([]float64, 0)
		}
		seriesTimestamps[seriesId] = append(seriesTimestamps[seriesId], item.ts)
		seriesValues[seriesId] = append(seriesValues[seriesId], item.v)
		seriesNamespace[seriesId] = item.series.namespace
	}

	// request (batch)
	request = types.WriteRequest{
		Series:        []types.WriteSeriesRequest{},
		SessionTicket: conn.getSessionTicket(),
	}

	// assemble request
	for seriesId, timestamps := range seriesTimestamps {
		request.Series = append(request.Series, types.WriteSeriesRequest{
			Times:  timestamps,
			Values: seriesValues[seriesId],
			SeriesIdentifier: types.SeriesIdentifier{
				Id:        seriesId,
				Namespace: seriesNamespace[seriesId],
			},
		})
	}
	return request, nil
}

func (batch *BatchWriter) Execute() (res WriteResult) {
	// get
	conn, err := batch.client.GetConnection()
	if err != nil {
		res.Error = err
		return
	}
	defer panicOnErrorClose(conn.Close)

	// to request
	request, err := batch.ToWriteRequest(conn)
	if err != nil {
		res.Error = err
		return
	}

	// execute
	var response *types.WriteResponse
	if err := conn.client.Call(types.EndpointWriter.String()+"."+types.MethodName, request, &response); err != nil {
		res.Error = err
		return
	}
	if response.Error != nil {
		res.Error = response.Error.Error()
		return
	}

	// validate num persisted
	res.NumPersisted = response.Num
	if res.NumPersisted != 1 {
		res.Error = errClientValidationMismatchSent
		return
	}

	return
}

func (batch *BatchWriter) AddToBatch(series *Series, ts uint64, v float64) error {
	if batch.items == nil {
		batch.items = make([]BatchItem, 0)
	}
	batch.items = append(batch.items, BatchItem{
		series: series,
		ts:     ts,
		v:      v,
	})
	return nil
}

type BatchItem struct {
	series *Series
	ts     uint64
	v      float64
}

func (client *Instance) NewBatchWriter() *BatchWriter {
	return &BatchWriter{
		client: client,
	}
}
