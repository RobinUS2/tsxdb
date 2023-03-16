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
		MaxSize:     50,
		New: func() interface{} {
			if client.closing {
				return nil
			}
			c, err := client.NewClient()
			if err != nil {
				log.Printf("init connection err %s", errors.Wrap(err, "failed to init new connection"))
				return nil
			}
			if client.opts.Debug {
				numCreated := atomic.LoadUint64(&client.connectionPool.numCreated)
				numDiscarded := atomic.LoadUint64(&client.connectionPool.numDiscarded)
				numConnections := numCreated - numDiscarded
				log.Printf("numConnections %d numCreated %d numDiscarded %d", numConnections, numCreated, numDiscarded)
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
	MaxSize     int // if > 0, will limit the amount that can be created
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
			p.pool <- p.runNew()
		}
	}
	return p
}

type GenericPool struct {
	numCreated   uint64
	numDiscarded uint64
	pool         chan interface{}
	New          func() interface{}
	opts         PoolOpts
}

func (p *GenericPool) runNew() interface{} {
	atomic.AddUint64(&p.numCreated, 1)
	return p.New()
}

func (p *GenericPool) Discard(_ interface{}) {
	atomic.AddUint64(&p.numDiscarded, 1)
}

func (p *GenericPool) Get() interface{} {
	// non-blocking from pool
	select {
	case v := <-p.pool:
		return v
	default:
		// unable to read directly from the channel, so pool is basically empty (all in-use)
		// limit?
		if p.opts.MaxSize > 0 {
			// create new ones again up until a certain maximum
			active := atomic.LoadUint64(&p.numCreated) - atomic.LoadUint64(&p.numDiscarded)
			available := uint64(p.opts.MaxSize) - active
			if available < 1 {
				// wait blocking until we get one back in the pool
				return <-p.pool
			}
		}
		return p.runNew()
	}
}

func (p *GenericPool) Put(v interface{}) {
	select {
	case p.pool <- v:
		// non blocking write
	default:
		// pool full, discard
		p.Discard(v)
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
