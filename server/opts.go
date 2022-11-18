package server

import (
	"github.com/RobinUS2/tsxdb/rpc"
)

type Opts struct {
	rpc.OptsConnection `yaml:"connection"`
	TelnetPort         int                 `yaml:"telnet_port"`
	TelnetHost         string              `yaml:"telnet_host"`
	Backends           []BackendOpts       `yaml:"backends"`
	BackendStrategy    BackendStrategyOpts `yaml:"backendStrategy"`
}

type BackendOpts struct {
	Type       string                 `yaml:"type"`       // e.g. memory, redis
	Identifier string                 `yaml:"identifier"` // unique name
	Metadata   bool                   `yaml:"metadata"`   // if true, this will store the metadata
	Options    map[string]interface{} `yaml:"options"`    // backend specific options
}

type BackendStrategyOpts struct {
	Type    string                 `yaml:"type"`    // e.g. simple
	Options map[string]interface{} `yaml:"options"` // strategy specific options
}

func NewOpts() *Opts {
	return &Opts{
		OptsConnection: rpc.NewOptsConnection(),
	}
}
