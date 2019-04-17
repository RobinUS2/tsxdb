package client

import (
	"errors"
	"fmt"
	"net/rpc"
	"sync"
	"time"
)

func (client *Instance) initConnectionPool() error {
	client.connectionPool = &sync.Pool{
		New: func() interface{} {
			if client.closing {
				return nil
			}
			c, err := client.NewConnection()
			//log.Println("new con")
			if err != nil {
				// @todo what to do?
				panic(err)
			}
			return c
		},
	}
	return nil
}

func (client *Instance) GetConnection() (*ManagedConnection, error) {
	conn := client.connectionPool.Get()
	if conn == nil {
		if client.closing {
			return nil, nil
		}
		return nil, errors.New("failed to obtain connection")
	}
	return conn.(*ManagedConnection), nil
}

func (client *Instance) NewConnection() (*ManagedConnection, error) {
	conn, err := rpc.Dial("tcp", client.opts.ListenHost+fmt.Sprintf(":%d", client.opts.ListenPort))
	if err != nil {
		return nil, err
	}
	return &ManagedConnection{
		client:  conn,
		pool:    client.connectionPool,
		created: time.Now().Unix(),
	}, nil
}

type ManagedConnection struct {
	client  *rpc.Client
	pool    *sync.Pool
	created int64
}

func (conn *ManagedConnection) Close() error {
	// keep alive?
	if time.Now().Unix()-conn.created >= 60 {
		// re-use
		conn.pool.Put(conn)
		return nil
	}

	// close
	if err := conn.client.Close(); err != nil {
		return err
	}
	return nil
}

func (client *Instance) Close() {
	client.closing = true
	for {
		conn, _ := client.GetConnection()
		if conn != nil {
			_ = conn.client.Close()
		} else {
			break
		}
	}
}
