package server

import (
	"errors"
	"fmt"
	"github.com/RobinUS2/tsxdb/server/backend"
	"github.com/RobinUS2/tsxdb/server/rollup"
	"github.com/RobinUS2/tsxdb/telnet"
	"log"
	"net"
	"net/rpc"
	"strings"
	"sync"
	"sync/atomic"
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

	sessionTokens    map[int][]byte // session id => secret
	sessionTokensMux sync.RWMutex

	rpcListener    net.Listener
	rpcListenerMux sync.RWMutex

	metaStore backend.IMetadata

	telnetServer *telnet.Instance

	// stats
	Stats
}

func (instance *Instance) MetaStore() backend.IMetadata {
	return instance.metaStore
}

type Stats struct {
	numCalls             uint64
	numValuesWritten     uint64
	numSeriesCreated     uint64
	numSeriesInitialised uint64
	numAuthentications   uint64
	numReads             uint64
}

func (instance *Instance) Statistics() Stats {
	return Stats{
		numCalls:             atomic.LoadUint64(&instance.numCalls),
		numValuesWritten:     atomic.LoadUint64(&instance.numValuesWritten),
		numSeriesCreated:     atomic.LoadUint64(&instance.numSeriesCreated),
		numSeriesInitialised: atomic.LoadUint64(&instance.numSeriesInitialised),
		numAuthentications:   atomic.LoadUint64(&instance.numAuthentications),
		numReads:             atomic.LoadUint64(&instance.numReads),
	}
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
		opts:          opts,
		rpc:           rpc.NewServer(),
		rollupReader:  rollup.NewReader(),
		sessionTokens: make(map[int][]byte),
		connections:   make(map[net.Addr]net.Conn),
	}
}

func (instance *Instance) Init() error {
	// register all endpoints
	endpointOpts := &EndpointOpts{server: instance}
	for _, endpoint := range endpoints {
		if err := endpoint.register(endpointOpts); err != nil {
			return err
		}
	}

	// default backend?
	if len(instance.opts.Backends) < 1 {
		// default backend memory
		log.Printf("WARN no backends defined, creating default non-persistent embedded memory backend")
		instance.opts.Backends = []BackendOpts{
			{
				Identifier: backend.DefaultIdentifier,
				Type:       backend.MemoryType.String(),
				Metadata:   true,
			},
		}
	}

	// create backends
	backends := make([]backend.IAbstractBackend, 0)
	for _, backendOpt := range instance.opts.Backends {
		b := backend.InstanceFactory(backendOpt.Type, backendOpt.Options)
		if b == nil {
			panic(fmt.Sprintf("failed to construct backend %+v", backendOpt))
		}
		backends = append(backends, b)
	}
	if len(backends) > 1 {
		return errors.New("no more than 1 backend supported for now")
	}

	// backend strategy
	if len(strings.TrimSpace(instance.opts.BackendStrategy.Type)) < 1 {
		instance.opts.BackendStrategy.Type = backend.SimpleStrategyType.String()
	}
	myStrategy := backend.StrategyInstanceFactory(instance.opts.BackendStrategy.Type, instance.opts.BackendStrategy.Options)
	myStrategy.SetBackends(backends)

	// backend selector
	instance.backendSelector = backend.NewSelector()
	if err := instance.backendSelector.AddStrategy(myStrategy); err != nil {
		return err
	}

	// must have auth
	if len(strings.TrimSpace(instance.opts.AuthToken)) < 1 {
		return errors.New("missing mandatory auth token option")
	}

	// metadata
	var metadataBackend backend.AbstractBackendWithMetadata
	for i, backendInstance := range backends {
		backendOpts := instance.opts.Backends[i]
		if !backendOpts.Metadata {
			continue
		}
		if typed, ok := backendInstance.(backend.AbstractBackendWithMetadata); ok {
			metadataBackend = typed
		} else {
			panic(fmt.Sprintf("backend %+v claims incorrectly to be able to store metadata", backendOpts))
		}
	}
	if metadataBackend == nil {
		return errors.New("no metadata backend defined")
	}
	instance.metaStore = backend.NewMetadata(metadataBackend)

	// link backends back to the system
	for _, backendInstance := range backends {
		backendInstance.SetReverseApi(instance.metaStore)
	}

	// init backends
	for _, backendInstance := range backends {
		if err := backendInstance.Init(); err != nil {
			return err
		}
	}

	// stats ticker
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			log.Printf("stats %+v", instance.Statistics())
		}
	}()
	return nil
}

func (instance *Instance) Start() (err error) {
	// catch runtime errors
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("server runtime error %s", r)
		}
	}()

	// start server
	if err := instance.StartListening(); err != nil {
		return err
	}

	// telnet server
	if instance.Opts().TelnetPort > 0 {
		telOpts := telnet.NewOpts()
		telOpts.Port = instance.Opts().TelnetPort
		telOpts.Host = instance.Opts().TelnetHost
		telOpts.AuthToken = instance.Opts().AuthToken
		telOpts.ServerHost = instance.Opts().ListenHost
		telOpts.ServerPort = instance.Opts().ListenPort
		instance.telnetServer = telnet.New(telOpts)
		go func() {
			err := instance.telnetServer.Listen()
			if err != nil {
				log.Printf("telnet failed to listen %s", err)
			}
		}()
	}

	return nil
}

func (instance *Instance) Shutdown() error {
	log.Println("shutting down")
	atomic.StoreInt32(&instance.shuttingDown, 1)

	// poll RPC listener shutdown
	if instance.RpcListener() != nil {
		// pending
		v := atomic.LoadInt64(&instance.pendingRequests)
		if v > 0 {
			// 50 x 100ms => 5 second max
			for i := 0; i < 50; i++ {
				time.Sleep(100 * time.Millisecond)
				v := atomic.LoadInt64(&instance.pendingRequests)
				if instance.RpcListener() == nil || v == 0 {
					break
				}
			}
		}
		// force shutdown
		if err := instance.closeRpc(); err != nil {
			return err
		}
	}

	// shutdown telnet
	if instance.telnetServer != nil {
		if err := instance.telnetServer.Shutdown(); err != nil {
			return err
		}
	}

	log.Println("shutdown complete")
	return nil
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
