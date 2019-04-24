package backend

import (
	"../../rpc/types"
)

type IMetadata interface {
	// @todo params, responses etc
	CreateOrUpdateSeries(*CreateSeries) *CreateSeriesResult // create/update new series (batch)
	SearchSeries()                                          // search one or multiple series by tags
	DeleteSeries()                                          // remove series (batch)
}

type CreateSeries struct {
	Series map[types.SeriesCreateIdentifier]types.SeriesMetadata
}

type CreateSeriesResult struct {
	Results map[types.SeriesCreateIdentifier]types.SeriesMetadataResponse
}
