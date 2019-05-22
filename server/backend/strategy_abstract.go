package backend

type AbstractStrategy interface {
	GetBackend() IAbstractBackend
}
