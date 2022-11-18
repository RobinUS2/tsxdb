package client

import (
	"sync"
	"sync/atomic"
	"time"
)

// automatically managed batch, can set a limit (number of items) and timeout (every x seconds)

const nanoToMs = 1000 * 1000

type AutoBatchWriter struct {
	batch    *BatchWriter
	batchMux sync.RWMutex

	client      *Instance
	batchSize   uint64
	timeoutMs   uint64
	flushMux    sync.RWMutex
	ticker      *time.Ticker
	postFlushFn func()

	// stats
	currentSize uint64
	flushCount  uint64
	lastFlush   uint64

	errors     chan error
	asyncFlush bool
}

func (instance *AutoBatchWriter) AsyncFlush() bool {
	return instance.asyncFlush
}

// subscribe to this channel to prevent panics in the ticker
func (instance *AutoBatchWriter) Errors(intOpts ...int) chan error {
	if instance.errors == nil {
		bufferSize := 1
		if len(intOpts) == 1 {
			bufferSize = intOpts[0]
		}
		instance.errors = make(chan error, bufferSize)
	}
	return instance.errors
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
	atomic.StoreUint64(&instance.currentSize, 0)
}

func (instance *AutoBatchWriter) Flush() error {
	// @todo this blocks AddToBatch until the flush is complete, make a copy of the batch and execute async instead
	// empty ?
	size := atomic.LoadUint64(&instance.currentSize)
	if size < 1 {
		return nil
	}

	// lock Flush
	instance.flushMux.Lock()
	defer instance.flushMux.Unlock()

	// lock the batch (for writing)
	instance.batchMux.Lock()
	defer instance.batchMux.Unlock()

	// double check
	if instance.batch.Size() < 1 {
		return nil
	}

	// copy batch
	itemsCopy := append([]BatchItem{}, instance.batch.items...)
	batchCopy := instance.client.NewBatchWriter()
	for _, itemCopy := range itemsCopy {
		if err := batchCopy.AddToBatch(itemCopy.series, itemCopy.ts, itemCopy.v); err != nil {
			return err
		}
	}

	// flush function
	executeFn := func() error {
		res := batchCopy.Execute()
		if res.Error != nil {
			return res.Error
		}

		// callback
		if instance.postFlushFn != nil {
			instance.postFlushFn()
		}
		return nil
	}
	if instance.asyncFlush {
		// async
		go func() {
			err := executeFn()
			if err != nil {
				instance.asyncError(err)
			}
		}()
	} else {
		// blocking
		if err := executeFn(); err != nil {
			return err
		}
	}

	// stats
	atomic.AddUint64(&instance.flushCount, 1)
	atomic.StoreUint64(&instance.lastFlush, nowMs())

	// new batch
	instance.initBatch()

	return nil
}

func (instance *AutoBatchWriter) AddToBatch(series *Series, ts uint64, v float64) error {
	instance.batchMux.Lock()
	if err := instance.batch.AddToBatch(series, ts, v); err != nil {
		instance.batchMux.Unlock()
		return err
	}
	instance.batchMux.Unlock()
	// check size, for auth Flush
	newSize := atomic.AddUint64(&instance.currentSize, 1)
	if newSize >= instance.batchSize {
		return instance.Flush()
	}
	return nil
}

func (instance *AutoBatchWriter) asyncError(err error) {
	if instance.errors == nil {
		panic(err)
	}
	instance.errors <- err
}

func (instance *AutoBatchWriter) startFlusher() {
	instance.ticker = time.NewTicker(time.Duration(instance.timeoutMs) * time.Millisecond / 10)
	go func() {
		for range instance.ticker.C {
			lastFlush := atomic.LoadUint64(&instance.lastFlush)
			ts := nowMs()
			deltaT := ts - lastFlush
			if deltaT < instance.timeoutMs {
				continue
			}
			if err := instance.Flush(); err != nil {
				instance.asyncError(err)
			}
		}
	}()
}

func nowMs() uint64 {
	return uint64(time.Now().UnixNano() / nanoToMs)
}

func (instance *AutoBatchWriter) Close() error {
	instance.ticker.Stop()
	return nil
}

type AutoBatchOpt interface {
	Apply(*AutoBatchWriter) error
}

type AutoBatchOptAsyncFlush struct {
	val bool
}

func (opt *AutoBatchOptAsyncFlush) Apply(w *AutoBatchWriter) error {
	w.asyncFlush = opt.val
	return nil
}

func NewAutoBatchOptAsyncFlush(val bool) *AutoBatchOptAsyncFlush {
	return &AutoBatchOptAsyncFlush{
		val: val,
	}
}

func (client *Instance) NewAutoBatchWriter(batchSize uint64, timeout time.Duration, opts ...AutoBatchOpt) *AutoBatchWriter {
	autoBatchWriter := &AutoBatchWriter{
		client:     client,
		batchSize:  batchSize,
		timeoutMs:  uint64(timeout.Nanoseconds() / nanoToMs),
		lastFlush:  nowMs(),
		errors:     nil,
		asyncFlush: true, // default true, since also ran in ticker, can't monitor errors anyway
	}
	for _, opt := range opts {
		if err := opt.Apply(autoBatchWriter); err != nil {
			panic(err)
		}
	}
	autoBatchWriter.initBatch()
	autoBatchWriter.startFlusher()
	return autoBatchWriter
}
