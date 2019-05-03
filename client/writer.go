package client

import (
	"errors"
	"github.com/RobinUS2/tsxdb/rpc/types"
)

var clientValidationErrMismatchSent = errors.New("mismatch between expected written values and received")

func (series *Series) Write(ts uint64, v float64) (res WriteResult) {
	// get
	conn, err := series.client.GetConnection()
	if err != nil {
		res.Error = err
		return
	}
	defer panicOnErrorClose(conn.Close)

	// init series
	if err := series.Init(conn); err != nil {
		res.Error = err
		return
	}

	// request (single)
	request := types.WriteRequest{
		Series: []types.WriteSeriesRequest{
			{
				Times:  []uint64{ts},
				Values: []float64{v},
				SeriesIdentifier: types.SeriesIdentifier{
					Id:        series.id,
					Namespace: series.namespace,
				},
			},
		},
		SessionTicket: conn.getSessionTicket(),
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
		res.Error = clientValidationErrMismatchSent
		return
	}

	return
}

// @todo batch write

type WriteResult struct {
	Error        error
	NumPersisted int
}
