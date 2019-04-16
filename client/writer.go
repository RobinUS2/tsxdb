package client

import (
	"../rpc/types"
	"errors"
)

var clientValidationErrMismatchSent = errors.New("mismatch between expected written values and received")

func (series Series) Write(ts uint64, v float64) (res WriteResult) {
	// request (single)
	request := types.WriteRequest{
		Times:  []uint64{ts},
		Values: []float64{v},
	}

	// get
	conn, err := series.client.GetConnection()
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
	if res.NumPersisted != len(request.Values) {
		res.Error = clientValidationErrMismatchSent
		return
	}

	// close
	if err := conn.client.Close(); err != nil {
		res.Error = err
		return
	}

	return
}

// @todo batch write

type WriteResult struct {
	Error        error
	NumPersisted int
}
