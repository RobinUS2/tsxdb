package client

import (
	"github.com/RobinUS2/tsxdb/rpc/types"
)

// @todo support batch read to amortize the overhead (network round-trip, authorization, etc) amongst many reads, probably just make it possible to chain multiple queries, the ReadRequest already supports it based on the arrays in there

// for now just a read single
func (series *Series) Read(q Query) (res QueryResult) {

	// get
	conn, err := series.client.GetConnection()
	if err != nil {
		res.Error = err
		return
	}
	defer panicOnErrorClose(conn.Close)

	// init series
	var seriesId uint64
	if seriesId, err = series.Init(conn); err != nil {
		res.Error = err
		return
	}
	// request (single)
	request := types.ReadRequest{
		Queries: []types.ReadSeriesRequest{
			{
				From: q.From,
				To:   q.To,
				SeriesIdentifier: types.SeriesIdentifier{
					Id:        seriesId,
					Namespace: series.namespace,
				},
			},
		},
		SessionTicket: conn.getSessionTicket(),
	}

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
	res.Results = response.Results[series.id]

	return
}
