package client

import (
	"github.com/RobinUS2/tsxdb/rpc/types"
	"sort"
)

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
	if err := series.Init(conn); err != nil {
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
					Id:        series.id,
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

type QueryResult struct {
	Error   error
	Results map[uint64]float64 // in random order due to Go map implementation, if you need sorted results call QueryResult.Iterator()
}

func (res QueryResult) Iterator() *QueryResultIterator {
	iter := &QueryResultIterator{
		results: &res,
		size:    len(res.Results),
	}
	iter.Reset()

	// sort
	sortedKeys := make([]uint64, iter.size)
	idx := 0
	for k := range iter.results.Results {
		sortedKeys[idx] = k
		idx++
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })
	iter.dataKeys = sortedKeys

	return iter
}

type QueryResultIterator struct {
	results  *QueryResult
	current  int
	size     int
	dataKeys []uint64
}

func (iter *QueryResultIterator) Reset() {
	iter.current = -1 // start before first, since you do for it.Next() { it.Value() }
}

func (iter *QueryResultIterator) Next() bool {
	iter.current++
	return iter.current < iter.size
}

func (iter *QueryResultIterator) Value() (uint64, float64) {
	timestamp := iter.dataKeys[iter.current]
	value := iter.results.Results[timestamp]
	return timestamp, value
}
