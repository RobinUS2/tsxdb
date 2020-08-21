package client

func (series *Series) applyOpts(opts []SeriesOpt) {
	series.metaMux.Lock()
	for _, opt := range opts {
		if err := opt.Apply(series); err != nil {
			panic(err)
		}
	}
	series.metaMux.Unlock()
}

type SeriesOpt interface {
	Apply(*Series) error
}
