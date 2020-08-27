package client

type Instance struct {
	opts           *Opts
	numConnections int64
	connectionPool *ConnectionPool
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
