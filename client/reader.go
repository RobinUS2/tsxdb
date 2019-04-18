package client

import (
	"../rpc/types"
)

func (series Series) Read(q Query) (res QueryResult) {
	// request (single)
	request := types.ReadRequest{
		From: q.From,
		To:   q.To,
		SeriesIdentifier: types.SeriesIdentifier{
			Id: series.id,
		},
	}
	// @todo inject session ticket

	// get
	conn, err := series.client.GetConnection()
	if err != nil {
		res.Error = err
		return
	}
	defer panicOnErrorClose(conn.Close)

	// execute
	var response *types.ReadResponse
	if err := conn.client.Call(types.EndpointReader.String()+"."+types.MethodName, request, &response); err != nil {
		res.Error = err
		return
	}
	if response.Error != nil {
		res.Error = response.Error.Error()
		return
	}
	res.Results = response.Results

	return
}

type QueryResult struct {
	Error   error
	Results map[uint64]float64
}
