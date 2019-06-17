package backend

var SimpleStrategyType = StrategyType("simple")

type SimpleStrategy struct {
	backend IAbstractBackend
}

func (strategy *SimpleStrategy) SetBackends(backend []IAbstractBackend) {
	if len(backend) != 1 {
		panic("simple strategy only supports 1 backend")
	}
	strategy.backend = backend[0]
}

func (strategy *SimpleStrategy) GetBackend() IAbstractBackend {
	return strategy.backend
}

func NewSimpleStrategy() AbstractStrategy {
	return &SimpleStrategy{}
}
