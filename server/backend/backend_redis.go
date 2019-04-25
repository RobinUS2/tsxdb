package backend

import (
	"errors"
)

type RedisBackend struct {
}

func (instance *RedisBackend) Type() TypeBackend {
	return TypeBackend("redis")
}

func (instance *RedisBackend) Write(context ContextWrite, timestamps []uint64, values []float64) error {
	return errors.New("not implemented")
}

func (instance *RedisBackend) Read(context ContextRead) (res ReadResult) {
	res.Error = errors.New("not implemented")

	return
}

func (instance *RedisBackend) CreateOrUpdateSeries(create *CreateSeries) (result *CreateSeriesResult) {
	return nil
}

func (instance *RedisBackend) SearchSeries(search *SearchSeries) (result *SearchSeriesResult) {
	return nil
}

func (instance *RedisBackend) DeleteSeries(ops *DeleteSeries) (result *DeleteSeriesResult) {
	return nil
}

func NewRedisBackend() *RedisBackend {
	return &RedisBackend{}
}
