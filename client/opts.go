package client

import (
	"github.com/RobinUS2/tsxdb/rpc"
)

type Opts struct {
	rpc.OptsConnection
	SeriesCacheSize int64
	EagerInitSeries bool // will load metadata on creation (async, instead of during flush, more equally spreading out load)
}

func NewOpts() *Opts {
	return &Opts{
		OptsConnection:  rpc.NewOptsConnection(),
		SeriesCacheSize: 100 * 1000, // by default keep 100K series metadata IDs in-memory
		EagerInitSeries: true,
	}
}
