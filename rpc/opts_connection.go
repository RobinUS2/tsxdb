package rpc

const DefaultListenPort = 1234
const DefaultListenHost = "127.0.0.1"

type OptsConnection struct {
	ListenPort int    `yaml:"listen_port"`
	ListenHost string `yaml:"listen_host"`
	AuthToken  string `yaml:"auth_token"`
}

func NewOptsConnection() OptsConnection {
	return OptsConnection{
		ListenPort: DefaultListenPort,
		ListenHost: DefaultListenHost,
	}
}
