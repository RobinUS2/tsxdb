package client

import (
	"github.com/RobinUS2/tsxdb/rpc"
)

type Opts struct {
	rpc.OptsConnection
	SeriesCacheSize int64
}

func NewOpts() *Opts {
	return &Opts{
		OptsConnection:  rpc.NewOptsConnection(),
		SeriesCacheSize: 1000 * 1000, // by default keep 1MM series metadata IDs in-memory
	}
}
