package client

type Instance struct {
	opts *Opts
}

func New(opts *Opts) *Instance {
	return &Instance{
		opts: opts,
	}
}

func DefaultClient() *Instance {
	return New(NewOpts())
}
