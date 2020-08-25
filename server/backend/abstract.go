package backend

type IAbstractBackend interface {
	Type() TypeBackend
	Write(context ContextWrite, timestamps []uint64, values []float64) error
	FlushPendingWrites(requestId RequestId) error
	Read(context ContextRead) ReadResult
	Init() error // should be called before first usage
	SetReverseApi(IReverseApi)
}

type AbstractBackend struct {
	reverseApi IReverseApi
}

func (a *AbstractBackend) ReverseApi() IReverseApi {
	return a.reverseApi
}

func (a *AbstractBackend) SetReverseApi(reverseApi IReverseApi) {
	a.reverseApi = reverseApi
}

// backend that supports both metadata and storage
type AbstractBackendWithMetadata interface {
	IAbstractBackend
	IMetadata
}

type TypeBackend string

func (t TypeBackend) String() string {
	return string(t)
}

const DefaultIdentifier = "default"
