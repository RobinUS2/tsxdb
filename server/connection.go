package server

import (
	"fmt"
	"net"
)

func (instance *Instance) StartListening() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", instance.opts.ListenPort))
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
				break
			}

			go instance.ServeConn(conn)
		}
	}()

	return nil
}

func (instance *Instance) ServeConn(conn net.Conn) {
	instance.rpc.ServeConn(conn)
}
