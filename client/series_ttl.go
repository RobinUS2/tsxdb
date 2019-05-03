package client

type SeriesTTL struct {
	ttl uint
}

func (opt SeriesTTL) Apply(series *Series) error {
	series.ttl = opt.ttl
	return nil
}

func NewSeriesTTL(ttl uint) *SeriesTTL {
	return &SeriesTTL{ttl: ttl}
}
