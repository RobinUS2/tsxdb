package server

import (
	"./backend"
)

func (instance *Instance) SelectBackend(context backend.ContextBackend) (backend.AbstractBackend, error) {
	selectedStrategy, err := instance.backendSelector.SelectStrategy(context)
	if err != nil {
		return nil, err
	}
	b := selectedStrategy.GetBackend()
	return b, nil
}
