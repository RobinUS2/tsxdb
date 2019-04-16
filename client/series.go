package client

type Series struct {
	client *Instance
}

type SeriesOpt struct {
}

func (client *Instance) Series(name string, opts ...SeriesOpt) *Series {
	return NewSeries(client)
}

func NewSeries(client *Instance) *Series {
	return &Series{
		client: client,
	}
}
