package backend

import (
	"../../rpc/types"
)

type IMetadata interface {
	// @todo params, responses etc
	CreateOrUpdateSeries(*CreateSeries) *CreateSeriesResult // create/update new series (batch)
	SearchSeries(*SearchSeries) *SearchSeriesResult         // search one or multiple series by tags
	DeleteSeries(*DeleteSeries) *DeleteSeriesResult         // remove series (batch)
}

type CreateSeries struct {
	Series map[types.SeriesCreateIdentifier]types.SeriesMetadata
	Error  error
}

type CreateSeriesResult struct {
	Results map[types.SeriesCreateIdentifier]types.SeriesMetadataResponse
}

type SearchSeries struct {
	SearchSeriesElement
}

type SearchSeriesResult struct {
	Series []types.SeriesIdentifier
	Error  error
}

type DeleteSeries struct {
	Series []types.SeriesIdentifier
}

type DeleteSeriesResult struct {
	Error error
}

type SearchSeriesElement struct {
	Namespace  int
	Name       string
	Tag        string
	Comparator SearchSeriesComparator
	And        []SearchSeriesElement
	Or         []SearchSeriesElement
}

type SearchSeriesComparator string

var SearchSeriesComparatorEquals SearchSeriesComparator = "EQUALS"
var SearchSeriesComparatorNot SearchSeriesComparator = "NOT"
