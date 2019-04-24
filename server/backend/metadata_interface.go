package backend

type IMetadata interface {
	// @todo params, responses etc
	CreateOrUpdateSeries() // create/update new series (batch)
	SearchSeries()         // search one or multiple series by tags
	DeleteSeries()         // remove series (batch)
}
