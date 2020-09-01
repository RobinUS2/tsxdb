package client

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Series struct {
	client    *Instance
	tags      []string
	namespace int
	id        uint64
	ttl       uint // in seconds, ~ 50.000 days
	name      string
	metaMux   sync.RWMutex

	initState    InitState
	initStateMux sync.RWMutex
	initMux      sync.Mutex
}

type InitState int64

const InitialState = 0
const TimeoutState = 1
const PanicState = 2
const SuccessState = 3
const ErrorState = 4

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

func (series *Series) InitState() InitState {
	series.initStateMux.RLock()
	state := series.initState
	series.initStateMux.RUnlock()
	return state
}

func (series *Series) SetInitState(state InitState) {
	series.initStateMux.Lock()
	series.initState = state
	series.initStateMux.Unlock()
}

func (client *Instance) Series(name string, opts ...SeriesOpt) *Series {
	s := NewSeries(name, client)
	isNew := s.Id() < 1
	if !isNew {
		return s
	}

	// initialise
	s.applyOpts(opts)

	if client.opts.EagerInitSeries {
		client.EagerInitSeries(s)
	}

	return s
}

const EagerSeriesInitTimeout = time.Second * 10

func (client *Instance) EagerInitSeries(series *Series) {
	// async to not block it, errors are ignored, since this is just a best effort, will be done (and error-ed) in write anyway later if retried
	go func() {
		// prevent a ton of concurrent inits, not good for opening lots of connections at once

		client.seriesPool.eagerInitMux.Lock()
		defer func() {
			client.seriesPool.eagerInitMux.Unlock()
		}()

		finished := make(chan bool, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("error initing series %v", r)
					series.SetInitState(PanicState)
				}
			}()
			if client.preEagerInitFn != nil {
				client.preEagerInitFn(series)
			}
			_, err := series.Create()
			if err != nil {
				log.Printf("error eager init for %s %s", series.Name(), err)
				series.SetInitState(ErrorState)
			} else {
				series.SetInitState(SuccessState)
			}
			finished <- true
		}()

		select {
		case <-finished:
			return
		case <-time.After(EagerSeriesInitTimeout):
			series.SetInitState(TimeoutState)
			return
		}
	}()
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
