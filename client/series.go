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
	metaMux   sync.RWMutex

	initMux sync.Mutex
}

func (series *Series) Id() uint64 {
	return atomic.LoadUint64(&series.id)
}

func (series *Series) Name() string {
	series.metaMux.RLock()
	v := series.name
	series.metaMux.RUnlock()
	return v
}

func (series *Series) TTL() uint {
	series.metaMux.RLock()
	v := series.ttl
	series.metaMux.RUnlock()
	return v
}

func (series *Series) Namespace() int {
	series.metaMux.RLock()
	v := series.namespace
	series.metaMux.RUnlock()
	return v
}

func (series *Series) Tags() []string {
	series.metaMux.RLock()
	v := series.tags
	series.metaMux.RUnlock()
	return v
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
