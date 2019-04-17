package backend

type AbstractStrategy interface {
	GetBackend() AbstractBackend
}
