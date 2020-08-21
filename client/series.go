package client

import (
	"sync"
	"sync/atomic"
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

func (series *Series) Id() uint64 {
	return atomic.LoadUint64(&series.id)
}

func (series *Series) Name() string {
	return series.name
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
	isNew := s.Id() < 1
	if !isNew {
		return s
	}

	// initialise
	s.applyOpts(opts)

	// eager init?
	if client.opts.EagerInitSeries {
		// async to not block it, errors are ignored, since this is just a best effort, will be done (and error-ed) in write anyway later if retried
		go func() {
			_, _ = s.Create()
		}()
	}

	return s
}

// Deprecated: use *Instance.Series() instead
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

	// set in pool
	client.seriesPool.Set(name, series)

	return series
}
