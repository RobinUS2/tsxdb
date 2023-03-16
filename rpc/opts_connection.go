package rpc

import "time"

const DefaultListenPort = 1234
const DefaultListenHost = "127.0.0.1"
const DefaultConnectTimeout = 10 * time.Second // setup a new connection to the server
const DefaultTimeout = 60 * time.Second        // timeout for a connection

type OptsConnection struct {
	ListenPort     int           `yaml:"listen_port"`
	ListenHost     string        `yaml:"listen_host"`
	AuthToken      string        `yaml:"auth_token"`
	ConnectTimeout time.Duration `yaml:"connection_timeout"`
	Debug          bool          `yaml:"debug"`
}

func NewOptsConnection() OptsConnection {
	return OptsConnection{
		ListenPort:     DefaultListenPort,
		ListenHost:     DefaultListenHost,
		ConnectTimeout: DefaultConnectTimeout,
		Debug:          true, // @todo set to false again
	}
}
