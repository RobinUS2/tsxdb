package client

type Series struct {
	client    *Instance
	tags      []string
	namespace int
}

func (series *Series) Namespace() int {
	return series.namespace
}

func (series Series) Tags() []string {
	return series.tags
}

func (client *Instance) Series(name string, opts ...SeriesOpt) *Series {
	s := NewSeries(client)
	s.applyOpts(opts)
	return s
}

func NewSeries(client *Instance) *Series {
	return &Series{
		client: client,
	}
}
