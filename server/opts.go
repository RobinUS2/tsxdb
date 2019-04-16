package server

import (
	"../rpc"
)

type Opts struct {
	rpc.OptsConnection
}

func NewOpts() *Opts {
	return &Opts{
		OptsConnection: rpc.NewOptsConnection(),
	}
}
