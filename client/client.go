package client

import "sync"

type Instance struct {
	opts           *Opts
	connectionPool *sync.Pool
	closing        bool
	seriesPool     *SeriesPool
}

func New(opts *Opts) *Instance {
	i := &Instance{
		opts:       opts,
		seriesPool: NewSeriesPool(opts),
	}
	if err := i.initConnectionPool(); err != nil {
		panic(err)
	}
	return i
}

func DefaultClient() *Instance {
	return New(NewOpts())
}
