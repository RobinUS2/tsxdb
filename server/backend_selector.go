package server

import (
	"github.com/RobinUS2/tsxdb/server/backend"
)

func (instance *Instance) SelectBackend(context backend.ContextBackend) (backend.IAbstractBackend, error) {
	selectedStrategy, err := instance.backendSelector.SelectStrategy(context)
	if err != nil {
		return nil, err
	}
	b := selectedStrategy.GetBackend()
	return b, nil
}
