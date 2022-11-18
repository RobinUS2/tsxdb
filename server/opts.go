package server

import (
	"github.com/RobinUS2/tsxdb/rpc"
)

type Opts struct {
	rpc.OptsConnection
}

func NewOpts() *Opts {
	return &Opts{
		OptsConnection: rpc.NewOptsConnection(),
	}
}
