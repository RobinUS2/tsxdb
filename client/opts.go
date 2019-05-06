package client

import (
	"github.com/Route42/tsxdb/rpc"
)

type Opts struct {
	rpc.OptsConnection
}

func NewOpts() *Opts {
	return &Opts{
		OptsConnection: rpc.NewOptsConnection(),
	}
}
