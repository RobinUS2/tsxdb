package server

import "sync/atomic"

type Stats struct {
	numCalls             uint64
	numValuesWritten     uint64
	numSeriesCreated     uint64
	numSeriesInitialised uint64
	numAuthentications   uint64
	numReads             uint64
}

func (s Stats) NumReads() uint64 {
	return s.numReads
}

func (s Stats) NumAuthentications() uint64 {
	return s.numAuthentications
}

func (s Stats) NumSeriesInitialised() uint64 {
	return s.numSeriesInitialised
}

func (s Stats) NumSeriesCreated() uint64 {
	return s.numSeriesCreated
}

func (s Stats) NumValuesWritten() uint64 {
	return s.numValuesWritten
}

func (s Stats) NumCalls() uint64 {
	return s.numCalls
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
