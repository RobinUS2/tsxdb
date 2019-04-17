package backend

type Selector struct {
	strategies []AbstractStrategy
}

func (selector *Selector) SelectStrategy(context ContextBackend) (AbstractStrategy, error) {
	// @todo proper implementation
	return selector.strategies[0], nil
}

func (selector *Selector) AddStrategy(strategy AbstractStrategy) error {
	selector.strategies = append(selector.strategies, strategy)
	return nil
}

func NewSelector() *Selector {
	return &Selector{
		strategies: make([]AbstractStrategy, 0),
	}
}
