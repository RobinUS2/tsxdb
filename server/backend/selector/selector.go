package selector

import (
	"../strategy"
)

type Selector struct {
	strategies []strategy.Abstract
}

func (selector *Selector) SelectStrategy(context Context) (strategy.Abstract, error) {
	// @todo proper implementation
	return selector.strategies[0], nil
}

func (selector *Selector) AddStrategy(strategy strategy.Abstract) error {
	selector.strategies = append(selector.strategies, strategy)
	return nil
}

func NewSelector() *Selector {
	return &Selector{
		strategies: make([]strategy.Abstract, 0),
	}
}

type Context struct {
	// details to determine context (e.g. series / metadata of series for persistence level)
}
