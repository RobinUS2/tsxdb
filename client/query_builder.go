package client

import (
	"errors"
	"math"
)

const QueryBuilderFromInf uint64 = 1
const QueryBuilderToInf uint64 = math.MaxUint64

type QueryBuilder struct {
	series *Series
	from   uint64
	to     uint64
}

func (series *Series) QueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		series: series,
	}
}

func (builder *QueryBuilder) From(from uint64) *QueryBuilder {
	builder.from = from
	return builder
}

func (builder *QueryBuilder) To(to uint64) *QueryBuilder {
	builder.to = to
	return builder
}

func (builder *QueryBuilder) IsValid() error {
	if builder.from == 0 || builder.to == 0 {
		return errors.New("missing time range")
	}
	return nil
}

func (builder *QueryBuilder) ToQuery() (*Query, error) {
	// validate
	if err := builder.IsValid(); err != nil {
		return nil, err
	}

	// query
	query := Query{
		Series: builder.series,
		From:   builder.from,
		To:     builder.to,
	}
	return &query, nil
}

func (builder *QueryBuilder) Execute() (res QueryResult) {
	// abstract via multi query builder
	multiQueryBuilder := builder.series.client.MultiQueryBuilder()
	err := multiQueryBuilder.AddQuery(builder)
	if err != nil {
		res.Error = err
		return
	}
	multiResult := multiQueryBuilder.Execute()
	if multiResult.Error != nil {
		res.Error = multiResult.Error
		return
	}
	// first result since we only add one query
	return multiResult.Results[0]
}

type Query struct {
	Series *Series
	From   uint64
	To     uint64
}
