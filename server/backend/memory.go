package backend

import (
	"log"
	"math/rand"
	"sync"
)

type MemoryBackend struct {
	data    map[int]map[uint64]map[float64]float64
	dataMux sync.RWMutex
}

func (instance *MemoryBackend) Type() TypeBackend {
	return TypeBackend("memory")
}

func (instance *MemoryBackend) Write(namespace int, series uint64, timestamp uint64, value float64) error {
	instance.dataMux.Lock()
	if _, found := instance.data[namespace]; !found {
		instance.data[namespace] = make(map[uint64]map[float64]float64)
	}
	if _, found := instance.data[namespace][series]; !found {
		instance.data[namespace][series] = make(map[float64]float64)
	}
	tsWithRand := float64(timestamp) + rand.Float64()
	instance.data[namespace][series][tsWithRand] = value
	log.Printf("%+v", instance.data)
	instance.dataMux.Unlock()
	return nil
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		data: make(map[int]map[uint64]map[float64]float64),
	}
}
