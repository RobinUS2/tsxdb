package tsxdb

import "github.com/Route42/tsxdb/client"

func NewClientOpts() *client.Opts {
	return client.NewOpts()
}

func NewClient(opts *client.Opts) *client.Instance {
	return client.New(opts)
}
