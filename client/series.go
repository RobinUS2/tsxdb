package client

import (
	"sync"
)

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

func (series *Series) Tags() []string {
	return series.tags
}

func (client *Instance) Series(name string, opts ...SeriesOpt) *Series {
	s := NewSeries(name, client)
	s.applyOpts(opts)
	return s
}

func NewSeries(name string, client *Instance) *Series {
	// automatic singleton to prevent connection re-init every time if user does not
	if series := client.seriesPool.Get(name); series != nil {
		return series
	}

	// create
	series := &Series{
		name:   name,
		client: client,
		id:     0, // will be populated on first usage
	}

	// eager init?
	if client.opts.EagerInitSeries {
		// async to not block it, errors are ignored, since this is just a best effort, will be done (and error-ed) in write anyway later if retried
		go func() {
			conn, err := client.GetConnection()
			if err != nil {
				return
			}
			_, _ = series.Init(conn)
		}()
	}

	// set in pool
	client.seriesPool.Set(name, series)

	return series
}
