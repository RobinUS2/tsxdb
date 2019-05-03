package client

import "sync"

type Series struct {
	client    *Instance
	tags      []string
	namespace int
	id        uint64
	ttl       uint // in seconds, ~ 50.000 days
	name      string

	initMux sync.Mutex
}

func (series *Series) TTL() uint {
	return series.ttl
}

func (series *Series) Namespace() int {
	return series.namespace
}

func (series Series) Tags() []string {
	return series.tags
}

func (client *Instance) Series(name string, opts ...SeriesOpt) *Series {
	s := NewSeries(name, client)
	s.applyOpts(opts)
	return s
}

func NewSeries(name string, client *Instance) *Series {
	return &Series{
		name:   name,
		client: client,
		id:     0, // will be populated on first usage
	}
}
