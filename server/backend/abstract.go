package backend

type AbstractBackend interface {
	Type() TypeBackend
	// @todo properties object
	Write(namespace int, series uint64, timestamp uint64, value float64) error
}

type TypeBackend string
