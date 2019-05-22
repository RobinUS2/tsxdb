package backend

type SimpleStrategy struct {
	backend IAbstractBackend
}

func (strategy *SimpleStrategy) GetBackend() IAbstractBackend {
	return strategy.backend
}

func NewSimpleStrategy(backend IAbstractBackend) AbstractStrategy {
	return &SimpleStrategy{
		backend: backend,
	}
}
