package client

func (series *Series) applyOpts(opts []SeriesOpt) {
	for _, opt := range opts {
		if err := opt.Apply(series); err != nil {
			panic(err)
		}
	}
}

type SeriesOpt interface {
	Apply(*Series) error
}
