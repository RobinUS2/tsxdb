package telnet

type Opts struct {
	Host      string
	Port      int
	AuthToken string
}

func NewOpts() *Opts {
	return &Opts{
		Host: "127.0.0.1", // default localhost for security
		Port: 5555,
	}
}
