package client

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/pkg/errors"
	"strings"
)

type MultiQueryBuilder struct {
	queries []Query
	client  *Instance
}

func (multi *MultiQueryBuilder) AddQuery(queryBuilder *QueryBuilder) error {
	query, err := queryBuilder.ToQuery()
	if err != nil {
		return err
	}

	seriesIds := make(map[uint64]bool)
	for _, query := range multi.queries {
		id := query.Series.Id()
		if seriesIds[id] {
			return fmt.Errorf("can only execute 1 query per series at this time") // needs refactoring of rpc results code which now returns the results by series id
		}
		seriesIds[id] = true
	}

	multi.queries = append(multi.queries, *query)
	return nil
}

// will return queries in the order as they are added via AddQuery
func (multi *MultiQueryBuilder) Execute() (res MultiQueryResult) {
	// get
	conn, err := multi.client.GetConnection()
	if err != nil {
		res.Error = errors.Wrap(err, "failed get connection")
		return
	}
	defer func() {
		if res.Error != nil && conn != nil {
			conn.Discard()
		}
		panicOnErrorClose(conn.Close)
	}()

	// batch request
	request := types.ReadRequest{
		Queries:       []types.ReadSeriesRequest{},
		SessionTicket: conn.getSessionTicket(),
	}

	// init series
	var querySeriesIdxMap = make(map[uint64]int)
	for idx, query := range multi.queries {
		// get series ID
		var seriesId uint64
		if seriesId, err = query.Series.Init(conn); err != nil {
			res.Error = err
			return
		}
		querySeriesIdxMap[seriesId] = idx

		// query
		queryRequest := types.ReadSeriesRequest{
			From: query.From,
			To:   query.To,
			SeriesIdentifier: types.SeriesIdentifier{
				Id:        seriesId,
				Namespace: query.Series.Namespace(),
			},
		}
		request.Queries = append(request.Queries, queryRequest)
	}

	// execute with retries
	var response *types.ReadResponse
	err = handleRetry(func() error {
		if err := conn.client.Call(types.EndpointReader.String()+"."+types.MethodName, request, &response); err != nil {
			return err
		}
		if response.Error != nil {
			if strings.Contains(response.Error.String(), types.RpcErrorNoDataFound.String()) {
				// not retryable if no data
				panic(response.Error.String())
			}
			return response.Error.Error()
		}
		return nil
	})
	if err != nil {
		res.Error = err
		return
	}

	// verify number of results equals what we requested
	if len(request.Queries) != len(response.Results) {
		res.Error = fmt.Errorf("expected %d results got %d", len(request.Queries), len(response.Results))
		return
	}

	// assign results to indexed slice
	res.Results = make([]QueryResult, len(request.Queries))
	for seriesId, results := range response.Results {
		idx, ok := querySeriesIdxMap[seriesId]
		if !ok {
			panic("missing series in map, should never happen, potential loss of metadata")
		}
		res.Results[idx] = QueryResult{
			Series:  multi.queries[idx].Series,
			Results: results,
			Error:   nil,
		}
	}

	return
}

func (client *Instance) MultiQueryBuilder() *MultiQueryBuilder {
	return newMultiQueryBuilder(client)
}

// Internal use only, use *Instance.MultiQueryBuilder() instead
func newMultiQueryBuilder(client *Instance) *MultiQueryBuilder {
	return &MultiQueryBuilder{
		client:  client,
		queries: make([]Query, 0),
	}
}
