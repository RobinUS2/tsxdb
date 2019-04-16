package client

import (
	"fmt"
	"net/rpc"
)

func (client *Instance) GetConnection() (*ManagedConnection, error) {
	conn, err := rpc.Dial("tcp", client.opts.ListenHost+fmt.Sprintf(":%d", client.opts.ListenPort))
	if err != nil {
		return nil, err
	}
	return &ManagedConnection{
		client: conn,
	}, nil
}

type ManagedConnection struct {
	client *rpc.Client
}
