package backend

// the backend implementation can callback into a limited set of APIs to achieve things
// for example if TTL expiry is implemented on-read it will have to instruct removal of that series
type IReverseApi interface {
	DeleteSeries(delete *DeleteSeries) *DeleteSeriesResult
}
