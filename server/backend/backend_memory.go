package backend

import (
	"errors"
	"log"
	"math/rand"
	"sync"
)

const maxPaddingSize = 0.1

type MemoryBackend struct {
	// @todo partition by timestamp!!!
	data    map[int]map[uint64]map[float64]float64
	dataMux sync.RWMutex
}

func (instance *MemoryBackend) Type() TypeBackend {
	return TypeBackend("memory")
}

func (instance *MemoryBackend) Write(context ContextWrite, timestamps []uint64, values []float64) error {
	if len(timestamps) != len(values) {
		return errors.New("mismatch pairs")
	}

	// obtain write lock
	instance.dataMux.Lock()

	// init maps
	instance.__notLockedInitMaps(context.Context, true)

	// execute writes
	for idx, timestamp := range timestamps {
		value := values[idx]

		// pad timestamp with random number to make sure we can have multiple per actual time "padded decimals"
		tsWithRand := float64(timestamp) + (rand.Float64() * maxPaddingSize)

		// write
		instance.data[context.Namespace][context.Series][tsWithRand] = value
	}

	// debug print
	log.Printf("%+v", instance.data)

	// unlock
	instance.dataMux.Unlock()

	return nil
}

func (instance *MemoryBackend) __notLockedInitMaps(context Context, autoCreate bool) (available bool) {
	if _, found := instance.data[context.Namespace]; !found {
		if !autoCreate {
			return false
		}
		instance.data[context.Namespace] = make(map[uint64]map[float64]float64)
	}
	if _, found := instance.data[context.Namespace][context.Series]; !found {
		if !autoCreate {
			return false
		}
		instance.data[context.Namespace][context.Series] = make(map[float64]float64)
	}
	return true
}

func (instance *MemoryBackend) Read(context ContextRead) (res ReadResult) {
	instance.dataMux.RLock()
	if !instance.__notLockedInitMaps(context.Context, false) {
		// not available in the store
		res.Error = errors.New("no data found")
		instance.dataMux.RUnlock()
		return
	}
	series := instance.data[context.Namespace][context.Series]
	instance.dataMux.RUnlock()

	// prune
	var pruned map[uint64]float64
	fromFloat := float64(context.From)
	toFloat := float64(context.To) + maxPaddingSize // add a bit here since that's the maximum value of the padding
	for ts, value := range series {
		if ts < fromFloat || ts > toFloat {
			continue
		}
		if pruned == nil {
			// lazy init map, since it could be very well that we have no data
			pruned = make(map[uint64]float64)
		}
		// truncate timestamp to get rid of the padded decimals
		pruned[uint64(ts)] = value
	}

	res.Results = pruned

	return
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		data: make(map[int]map[uint64]map[float64]float64),
	}
}
