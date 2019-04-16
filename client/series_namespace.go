package client

type SeriesNamespace struct {
	namespace int
}

func (opt SeriesNamespace) Apply(series *Series) error {
	series.namespace = opt.namespace
	return nil
}

func NewSeriesNamespace(namespace int) *SeriesNamespace {
	return &SeriesNamespace{namespace: namespace}
}
