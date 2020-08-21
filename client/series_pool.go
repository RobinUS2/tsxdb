package client

import (
	"github.com/karlseguin/ccache/v2"
	"sync/atomic"
	"time"
)

// series cache, makes sure initialised series are reused

type SeriesPool struct {
	hits uint64
	lru  *ccache.Cache
}

func (pool *SeriesPool) Hits() uint64 {
	return atomic.LoadUint64(&pool.hits)
}

func (pool *SeriesPool) Count() int {
	return pool.lru.ItemCount()
}

func (pool *SeriesPool) Get(name string) *Series {
	item := pool.lru.Get(name)
	if item == nil {
		return nil
	}
	if item.Expired() {
		return nil
	}
	atomic.AddUint64(&pool.hits, 1)
	return item.Value().(*Series)
}

func (pool *SeriesPool) Set(name string, value *Series) {
	pool.lru.Set(name, value, 24*time.Hour)
}

func NewSeriesPool(clientOpts *Opts) *SeriesPool {
	return &SeriesPool{
		lru: ccache.New(ccache.Configure().MaxSize(clientOpts.SeriesCacheSize)),
	}
}

// Deprecated: only use for testing
func (client *Instance) SeriesPool() *SeriesPool {
	return client.seriesPool
}
