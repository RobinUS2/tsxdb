package client

type SeriesTTL struct {
	ttl uint
}

func (opt SeriesTTL) Apply(series *Series) error {
	series.ttl = opt.ttl
	return nil
}

// expires the whole series after creation.
// e.g. setting this to 86400 the whole series will be removed after 1 day,
// regardless of additional timeseries values being added
func NewSeriesTTL(ttl uint) *SeriesTTL {
	return &SeriesTTL{ttl: ttl}
}
