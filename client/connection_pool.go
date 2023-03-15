package client

import (
	"github.com/pkg/errors"
	"log"
	"sync/atomic"
)

type ConnectionPool struct {
	GenericPool
}

func (client *Instance) NewConnectionPool() *ConnectionPool {
	genericPool := NewGenericPool(PoolOpts{
		Size:        10,
		PreWarmSize: 10,
		New: func() interface{} {
			if client.closing {
				return nil
			}
			c, err := client.NewClient()
			if err != nil {
				// is dealt with in (client *Instance) GetConnection() (*ManagedConnection, error)
				panic(errors.Wrap(err, "failed to init new connection"))
			}
			numConnections := atomic.AddInt64(&client.numConnections, 1)
			const numConnectionsWarningThreshold = 100
			if client.opts.Debug && numConnections > numConnectionsWarningThreshold {
				log.Printf("numConnections %d", numConnections)
			}
			return c
		},
	})
	p := &ConnectionPool{
		GenericPool: genericPool,
	}
	return p
}

type PoolOpts struct {
	Size        int
	PreWarmSize int
	New         func() interface{}
}

func NewGenericPool(opts PoolOpts) GenericPool {
	if opts.Size < 1 {
		panic("pool must have size")
	}
	if opts.New == nil {
		panic("missing new")
	}
	p := GenericPool{
		opts: opts,
		New:  opts.New,
		pool: make(chan interface{}, opts.Size),
	}
	if opts.PreWarmSize > 0 {
		for i := 0; i < opts.PreWarmSize; i++ {
			p.pool <- p.New()
		}
	}
	return p
}

type GenericPool struct {
	pool chan interface{}
	New  func() interface{}
	opts PoolOpts
}

func (p *GenericPool) Get() interface{} {
	// non-blocking from pool
	select {
	case v := <-p.pool:
		return v
		//default:
		//	return p.New()
	}
}

func (p *GenericPool) Put(v interface{}) {
	select {
	case p.pool <- v:
		// non blocking write
	default:
		// pool full, discard
		if closeable, ok := v.(IClosablePoolValue); ok {
			closeable.DiscardPool()
		}
	}
}

func (p *GenericPool) Size() int {
	return len(p.pool)
}

func (p *GenericPool) Capacity() int {
	return cap(p.pool)
}

type IPool interface {
	Get() interface{}
	Put(interface{})
	Capacity() int
	Size() int
}

type IClosablePoolValue interface {
	DiscardPool()
}
