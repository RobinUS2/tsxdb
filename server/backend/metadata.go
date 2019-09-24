package backend

type Metadata struct {
	backend AbstractBackendWithMetadata
}

func (meta *Metadata) CreateOrUpdateSeries(create *CreateSeries) *CreateSeriesResult {
	return meta.backend.CreateOrUpdateSeries(create)
}

func (meta *Metadata) SearchSeries(search *SearchSeries) *SearchSeriesResult {
	return meta.backend.SearchSeries(search)
}

func (meta *Metadata) DeleteSeries(delete *DeleteSeries) *DeleteSeriesResult {
	return meta.backend.DeleteSeries(delete)
}

func (meta *Metadata) Clear() error {
	return meta.backend.Clear()
}

func NewMetadata(backend AbstractBackendWithMetadata) *Metadata {
	return &Metadata{
		backend: backend,
	}
}
