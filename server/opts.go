package server

import (
	"errors"
	"github.com/RobinUS2/tsxdb/rpc"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Opts struct {
	rpc.OptsConnection `yaml:"connection"`
	TelnetPort         int `yaml:"telnet_port"`
}

func (opts *Opts) ReadYamlFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if b == nil || len(b) < 1 {
		return errors.New("missing configuration data")
	}
	if err := yaml.Unmarshal(b, &opts); err != nil {
		return err
	}
	return nil
}

func NewOpts() *Opts {
	return &Opts{
		OptsConnection: rpc.NewOptsConnection(),
	}
}
