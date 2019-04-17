package backend

type SimpleStrategy struct {
	backend AbstractBackend
}

func (strategy *SimpleStrategy) GetBackend() AbstractBackend {
	return strategy.backend
}

func NewSimpleStrategy(backend AbstractBackend) AbstractStrategy {
	return &SimpleStrategy{
		backend: backend,
	}
}
