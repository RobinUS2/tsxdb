package backend_test

import (
	"../backend"
	"testing"
)

func TestNewRedisBackendSingleConnection(t *testing.T) {
	opts := &backend.RedisOpts{
		ConnectionDetails: map[backend.Namespace]backend.RedisConnectionDetails{
			backend.RedisDefaultConnectionNamespace: {
				Addr: "localhost",
				Port: 6379,
			},
		},
	}
	b := backend.NewRedisBackend(opts)
	if b == nil {
		t.Error()
	}
	if err := b.Init(); err != nil {
		t.Error(err)
	}
}

func TestNewRedisBackendMultiConnection(t *testing.T) {
	opts := &backend.RedisOpts{
		ConnectionDetails: map[backend.Namespace]backend.RedisConnectionDetails{
			backend.RedisDefaultConnectionNamespace: {
				Addr: "localhost",
				Port: 6379,
			},
			backend.Namespace(5): {
				Addr: "localhost",
				Port: 6379,
			},
		},
	}
	b := backend.NewRedisBackend(opts)
	if b == nil {
		t.Error()
	}
	if err := b.Init(); err != nil {
		t.Error(err)
	}
}
