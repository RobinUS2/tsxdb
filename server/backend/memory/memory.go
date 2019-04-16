package memory

import (
	"../../backend"
)

type Instance struct {
}

func (instance *Instance) Type() backend.Type {
	return backend.Type("memory")
}

func NewInstance() *Instance {
	return &Instance{}
}
