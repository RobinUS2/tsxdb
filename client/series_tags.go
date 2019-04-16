package client

type SeriesTags struct {
	tags []string
}

func (opt SeriesTags) Apply(series *Series) error {
	if series.tags == nil {
		series.tags = make([]string, 0)
	}
	series.tags = append(series.tags, opt.tags...)
	return nil
}

func NewSeriesTags(tags ...string) *SeriesTags {
	return &SeriesTags{tags: tags}
}
