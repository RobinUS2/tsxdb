package backend_test

import (
	"../backend"
	"testing"
	"time"
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

	// simple write
	now := uint64(time.Now().Unix() * 1000)
	{
		if err := b.Write(backend.ContextWrite{
			Context: backend.Context{
				Namespace: 5,
				Series:    1,
			},
		}, []uint64{now}, []float64{1.2}); err != nil {
			t.Error(err)
		}
	}

	// simple read
	{
		res := b.Read(backend.ContextRead{
			Context: backend.Context{
				Series:    1,
				Namespace: 5,
			},
			From: now - (86400 * 1000 * 2),
			To:   now + (86400 * 1000 * 1),
		})
		if res.Error != nil {
			t.Error(res.Error)
		}
	}
}
