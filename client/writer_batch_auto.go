package client

import (
	"sync"
	"sync/atomic"
	"time"
)

// automatically managed batch, can set a limit (number of items) and timeout (every x seconds)

type AutoBatchWriter struct {
	batch       *BatchWriter
	client      *Instance
	batchSize   uint64
	timeout     time.Duration
	flushMux    sync.RWMutex
	ticker      *time.Ticker
	postFlushFn func()

	// stats
	currentSize uint64
	flushCount  uint64
	lastFlush   uint64
}

func (instance *AutoBatchWriter) SetPostFlushFn(postFlushFn func()) {
	instance.postFlushFn = postFlushFn
}

func (instance *AutoBatchWriter) FlushCount() uint64 {
	return atomic.LoadUint64(&instance.flushCount)
}

// unsafe, not locked
func (instance *AutoBatchWriter) initBatch() {
	instance.batch = instance.client.NewBatchWriter()
}

func (instance *AutoBatchWriter) flush() error {
	// empty ?
	instance.flushMux.RLock()
	if instance.batch.Size() < 1 {
		instance.flushMux.RUnlock()
		return nil
	}
	instance.flushMux.RUnlock()

	// lock
	instance.flushMux.Lock()
	defer instance.flushMux.Unlock()

	// execute
	res := instance.batch.Execute()
	if res.Error != nil {
		return res.Error
	}

	// stats
	atomic.AddUint64(&instance.flushCount, 1)
	atomic.StoreUint64(&instance.lastFlush, uint64(time.Now().UnixNano()))

	// new batch
	instance.initBatch()

	// callback
	if instance.postFlushFn != nil {
		instance.postFlushFn()
	}

	return nil
}

func (instance *AutoBatchWriter) AddToBatch(series *Series, ts uint64, v float64) error {
	if err := instance.batch.AddToBatch(series, ts, v); err != nil {
		return err
	}
	// check size, for auth flush
	newSize := atomic.AddUint64(&instance.currentSize, 1)
	if newSize >= instance.batchSize {
		return instance.flush()
	}
	return nil
}

func (instance *AutoBatchWriter) startFlusher() {
	instance.ticker = time.NewTicker(instance.timeout / 10)
	go func() {
		for range instance.ticker.C {
			// @todo check last flush time
			if err := instance.flush(); err != nil {
				panic(err)
			}
		}
	}()
}

func (instance *AutoBatchWriter) Close() error {
	instance.ticker.Stop()
	return nil
}

func (client *Instance) NewAutoBatchWriter(batchSize uint64, timeout time.Duration) *AutoBatchWriter {
	autoBatchWriter := &AutoBatchWriter{
		client:    client,
		batchSize: batchSize,
		timeout:   timeout,
	}
	autoBatchWriter.initBatch()
	autoBatchWriter.startFlusher()
	return autoBatchWriter
}
