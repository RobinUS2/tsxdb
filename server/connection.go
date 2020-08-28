package server

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Connections struct {
	pendingRequests int64

	connections      map[net.Addr]net.Conn
	connectionsMux   sync.RWMutex
	connectionTicker *time.Ticker

	expireSlots    map[FutureUnixTime][]net.Conn // unix timestamp in future -> connection
	expireSlotsMux sync.RWMutex
}

const ConnectionTimeout = rpc.DefaultTimeout

func (instance *Instance) StartListening() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", instance.opts.ListenPort))
	if err != nil {
		return err
	}
	instance.SetRpcListener(listener)

	go func() {
		for {
			conn, err := listener.Accept()
			isShuttingDown := atomic.LoadInt32(&instance.shuttingDown) == 1
			if err != nil {
				if !isShuttingDown {
					log.Printf("%s", err)
				}
				break
			}

			go instance.ServeConn(conn)

			if isShuttingDown {
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
	listener := instance.RpcListener()
	if listener != nil {
		if err := listener.Close(); err != nil {
			return err
		}
		instance.SetRpcListener(nil)
	}
	return nil
}

func (instance *Instance) ServeConn(conn net.Conn) {
	// register connection
	instance.RegisterConn(conn)
	atomic.AddInt64(&instance.pendingRequests, 1)

	// buffered writer
	srv := rpc.NewGobServerCodec(conn)

	// serve
	instance.rpc.ServeCodec(srv)

	// unregister
	atomic.AddInt64(&instance.pendingRequests, -1)
}

func (instance *Instance) RegisterConn(conn net.Conn) {
	instance.connectionsMux.Lock()
	instance.connections[conn.RemoteAddr()] = conn
	instance.connectionsMux.Unlock()

	// register for future time in expire via ticker
	nowUnix := time.Now().Unix()
	expireTs := FutureUnixTime(nowUnix + int64(ConnectionTimeout.Seconds()) + 1)
	instance.Connections.expireSlotsMux.Lock()
	if instance.Connections.expireSlots[expireTs] == nil {
		instance.Connections.expireSlots[expireTs] = make([]net.Conn, 0)
	}
	instance.Connections.expireSlots[expireTs] = append(instance.Connections.expireSlots[expireTs], conn)
	instance.Connections.expireSlotsMux.Unlock()
}

func (instance *Instance) RemoveConn(conn net.Conn) {
	instance.connectionsMux.Lock()
	delete(instance.connections, conn.RemoteAddr())
	instance.connectionsMux.Unlock()
}

func (instance *Instance) connectionExpire() int {
	nowUnix := time.Now().Unix()

	// scan for expired slots
	expiredSlots := make([]FutureUnixTime, 0)
	instance.Connections.expireSlotsMux.RLock()
	for ts := range instance.Connections.expireSlots {
		if ts >= FutureUnixTime(nowUnix) {
			// not yet expired
			continue
		}
		expiredSlots = append(expiredSlots, ts)
	}
	instance.Connections.expireSlotsMux.RUnlock()

	if len(expiredSlots) < 1 {
		// nothing to expire
		return 0
	}

	// remove all expired tokens
	numDeleted := 0
	for _, expiredSlot := range expiredSlots {
		connections, found := instance.Connections.expireSlots[expiredSlot]
		if !found {
			continue
		}
		for _, conn := range connections {
			// async close, since involves IO, can be blocking
			go func(conn net.Conn) {
				_ = conn.Close()
				instance.RemoveConn(conn)
			}(conn)
			numDeleted++
		}
	}
	log.Printf("%d expired", numDeleted)

	return numDeleted
}

func NewConnections() Connections {
	return Connections{
		connections: make(map[net.Addr]net.Conn),
	}
}
