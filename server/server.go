package server

import (
	"./backend"
	"./rollup"
	"errors"
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
	rpcListener     net.Listener
	shuttingDown    int32 // set to true during shutdown

	pendingRequests int64

	sessionTokens    map[int][]byte // session id => secret
	sessionTokensMux sync.RWMutex
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

	// testing backend strategy in memory
	// @todo from config
	instance.backendSelector = backend.NewSelector()
	myStrategy := backend.NewSimpleStrategy(backend.NewMemoryBackend())
	if err := instance.backendSelector.AddStrategy(myStrategy); err != nil {
		return err
	}

	// must have auth
	if len(strings.TrimSpace(instance.opts.AuthToken)) < 1 {
		return errors.New("missing mandatory auth token option")
	}

	return nil
}

func (instance *Instance) Start() error {
	if err := instance.StartListening(); err != nil {
		return err
	}
	return nil
}

func (instance *Instance) Shutdown() error {
	log.Println("shutting down")
	atomic.StoreInt32(&instance.shuttingDown, 1)

	// poll RPC listener shutdown
	if instance.rpcListener != nil {
		// pending
		v := atomic.LoadInt64(&instance.pendingRequests)
		if v > 0 {
			// 50 x 100ms => 5 second max
			for i := 0; i < 50; i++ {
				time.Sleep(100 * time.Millisecond)
				v := atomic.LoadInt64(&instance.pendingRequests)
				if instance.rpcListener == nil || v == 0 {
					break
				}
			}
		}
		// force shutdown
		if err := instance.closeRpc(); err != nil {
			return err
		}
	}

	log.Println("shutdown complete")
	return nil
}
