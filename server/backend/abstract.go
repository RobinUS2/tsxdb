package backend

type AbstractBackend interface {
	Type() TypeBackend
	Write(context ContextWrite, timestamps []uint64, values []float64) error
	Read(context ContextRead) ReadResult
	Init() error // should be called before first usage
}

// backend that supports both metadata and storage
type AbstractBackendWithMetadata interface {
	AbstractBackend
	IMetadata
}

type TypeBackend string
