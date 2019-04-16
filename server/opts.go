package server

type Opts struct {
	ListenPort int
}

func NewOpts() *Opts {
	return &Opts{
		ListenPort: 1234,
	}
}
