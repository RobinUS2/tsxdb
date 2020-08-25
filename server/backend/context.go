package backend

type Context struct {
	// details to determine context (e.g. series / metadata of series for persistence level)
	Series    uint64
	Namespace int
	RequestId RequestId
}

type ContextBackend struct {
	Context
}

type ContextWrite struct {
	Context
}

type ContextRead struct {
	Context
	From uint64
	To   uint64
	// @todo rollup and such
}
