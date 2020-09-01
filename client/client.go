package client

type Instance struct {
	opts           *Opts
	numConnections int64
	connectionPool *ConnectionPool
	closing        bool
	seriesPool     *SeriesPool

	*EagerInitSeriesHelper
}

type EagerInitSeriesHelper struct {
	preEagerInitFn func(series *Series)
}

func (helper *EagerInitSeriesHelper) SetPreEagerInitFn(f func(series *Series)) {
	helper.preEagerInitFn = f
}

func New(opts *Opts) *Instance {
	i := &Instance{
		opts:                  opts,
		seriesPool:            NewSeriesPool(opts),
		EagerInitSeriesHelper: &EagerInitSeriesHelper{},
	}
	if err := i.initConnectionPool(); err != nil {
		panic(err)
	}
	return i
}

func DefaultClient() *Instance {
	return New(NewOpts())
}
