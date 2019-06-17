package backend

type AbstractStrategy interface {
	GetBackend() IAbstractBackend
	SetBackends([]IAbstractBackend)
}

type StrategyType string

func (t StrategyType) String() string {
	return string(t)
}
