package client

func (series *Series) applyOpts(opts []SeriesOpt) {
	series.initMux.Lock()
	for _, opt := range opts {
		if err := opt.Apply(series); err != nil {
			panic(err)
		}
	}
	series.initMux.Unlock()
}

type SeriesOpt interface {
	Apply(*Series) error
}
