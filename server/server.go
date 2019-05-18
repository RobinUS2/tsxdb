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
	rpcListener     net.Listener
	shuttingDown    int32 // set to true during shutdown

	pendingRequests int64

	sessionTokens    map[int][]byte // session id => secret
	sessionTokensMux sync.RWMutex

	metaStore backend.IMetadata
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
	myBackend := backend.NewMemoryBackend()
	if err := myBackend.Init(); err != nil {
		return err
	}
	myStrategy := backend.NewSimpleStrategy(myBackend)
	if err := instance.backendSelector.AddStrategy(myStrategy); err != nil {
		return err
	}

	// must have auth
	if len(strings.TrimSpace(instance.opts.AuthToken)) < 1 {
		return errors.New("missing mandatory auth token option")
	}

	// metadata
	instance.metaStore = backend.NewMetadata(myBackend)

	return nil
}

func (instance *Instance) Start() (err error) {
	// catch runtime errors
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("server runtime error %s", r))
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
		t := telnet.New(telOpts)
		go func() {
			err := t.Listen()
			if err != nil {
				panic(err)
			}
		}()
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
