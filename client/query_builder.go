package client

import "errors"

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

func (builder *QueryBuilder) Execute() (res QueryResult) {
	// @todo implement
	if builder.from == 0 || builder.to == 0 {
		res.Error = errors.New("missing time range")
		return
	}

	// query
	query := Query{
		From: builder.from,
		To:   builder.to,
	}
	return builder.series.Read(query)
}

type Query struct {
	From uint64
	To   uint64
}
