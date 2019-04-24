package backend

type Metadata struct {
}

func (meta *Metadata) CreateOrUpdateSeries(*CreateSeries) *CreateSeriesResult {
	return nil
}

func (meta *Metadata) SearchSeries() {

}

func (meta *Metadata) DeleteSeries() {

}

func NewMetadata() *Metadata {
	return &Metadata{}
}
