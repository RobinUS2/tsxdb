package client

import (
	"../rpc/types"
	"log"
)

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
	log.Printf("sending %+v", request)
	var response *types.WriteResponse
	if err := conn.client.Call(types.EndpointWriter.String()+".Execute", request, &response); err != nil {
		res.Error = err
		return
	}
	log.Printf("received %+v", response)

	// close
	if err := conn.client.Close(); err != nil {
		res.Error = err
		return
	}

	return
}

// @todo batch write

type WriteResult struct {
	Error error
}
