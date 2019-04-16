package strategy

import (
	"../../backend"
)

type SimpleStrategy struct {
	backend backend.Abstract
}

func (strategy *SimpleStrategy) GetBackend() backend.Abstract {
	return strategy.backend
}

func NewSimpleStrategy(backend backend.Abstract) *SimpleStrategy {
	return &SimpleStrategy{
		backend: backend,
	}
}
