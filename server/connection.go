package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func (instance *Instance) StartListening() error {
	var err error
	instance.rpcListener, err = net.Listen("tcp", fmt.Sprintf(":%d", instance.opts.ListenPort))
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := instance.rpcListener.Accept()
			if err != nil {
				if !instance.shuttingDown {
					log.Printf("%s", err)
				}
				break
			}

			go instance.ServeConn(conn)

			if instance.shuttingDown {
				break
			}
		}

		// close
		if err := instance.closeRpc(); err != nil {
			panic(err)
		}
	}()

	return nil
}

var closeMux sync.RWMutex

func (instance *Instance) closeRpc() error {
	closeMux.Lock()
	defer closeMux.Unlock()
	if instance.rpcListener != nil {
		if err := instance.rpcListener.Close(); err != nil {
			return err
		}
		instance.rpcListener = nil
	}
	return nil
}

func (instance *Instance) ServeConn(conn net.Conn) {
	// @todo check blocked
	atomic.AddInt64(&instance.pendingRequests, 1)
	log.Printf("connection from %v", conn.RemoteAddr())
	instance.rpc.ServeConn(conn)
	atomic.AddInt64(&instance.pendingRequests, -1)

	// auth timeout
	go func() {
		time.Sleep(100 * time.Millisecond)
		log.Printf("check auth from %v", conn.RemoteAddr())
		// @todo really check
		// @todo check errors, block host from connection
		_ = conn.Close()
	}()
}
