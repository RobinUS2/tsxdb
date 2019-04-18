package rpc

const DefaultListenPort = 1234
const DefaultListenHost = "127.0.0.1"

type OptsConnection struct {
	ListenPort int
	ListenHost string
	AuthToken  string
}

func NewOptsConnection() OptsConnection {
	return OptsConnection{
		ListenPort: DefaultListenPort,
		ListenHost: DefaultListenHost,
	}
}
