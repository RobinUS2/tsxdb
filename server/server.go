package server

import (
	"github.com/RobinUS2/tsxdb/server/backend"
	"github.com/RobinUS2/tsxdb/server/rollup"
	"github.com/RobinUS2/tsxdb/telnet"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Instance struct {
	opts            *Opts
	rpc             *rpc.Server
	backendSelector *backend.Selector
	rollupReader    *rollup.Reader
	shuttingDown    int32 // set to true during shutdown

	pendingRequests int64

	connections    map[net.Addr]net.Conn
	connectionsMux sync.RWMutex

	Sessions

	rpcListener    net.Listener
	rpcListenerMux sync.RWMutex

	metaStore backend.IMetadata

	telnetServer *telnet.Instance

	// stats
	Stats
	statsTicker *time.Ticker
}

func (instance *Instance) MetaStore() backend.IMetadata {
	return instance.metaStore
}

func (instance *Instance) RpcListener() net.Listener {
	instance.rpcListenerMux.RLock()
	x := instance.rpcListener
	instance.rpcListenerMux.RUnlock()
	return x
}

func (instance *Instance) SetRpcListener(rpcListener net.Listener) {
	instance.rpcListenerMux.Lock()
	instance.rpcListener = rpcListener
	instance.rpcListenerMux.Unlock()
}

func (instance *Instance) Opts() *Opts {
	return instance.opts
}

func New(opts *Opts) *Instance {
	return &Instance{
		opts:         opts,
		rpc:          rpc.NewServer(),
		rollupReader: rollup.NewReader(),
		Sessions:     NewSessions(),
		connections:  make(map[net.Addr]net.Conn),
	}
}

func (instance *Instance) RegisterConn(conn net.Conn) {
	instance.connectionsMux.Lock()
	instance.connections[conn.RemoteAddr()] = conn
	instance.connectionsMux.Unlock()
}

func (instance *Instance) RemoveConn(conn net.Conn) {
	instance.connectionsMux.Lock()
	delete(instance.connections, conn.RemoteAddr())
	instance.connectionsMux.Unlock()
}
