package backend

type AbstractBackend interface {
	Type() TypeBackend
	Write(context ContextWrite, timestamps []uint64, values []float64) error
	Read(context ContextRead) ReadResult
}

type TypeBackend string
