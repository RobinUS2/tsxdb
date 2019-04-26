package backend_test

import (
	"../backend"
	"math/rand"
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
	const seriesId = 1
	now := uint64(time.Now().Unix() * 1000)
	writeVal := rand.Float64()
	{
		if err := b.Write(backend.ContextWrite{
			Context: backend.Context{
				Namespace: 5,
				Series:    seriesId,
			},
		}, []uint64{now}, []float64{writeVal}); err != nil {
			t.Error(err)
		}
	}

	// simple read
	{
		res := b.Read(backend.ContextRead{
			Context: backend.Context{
				Series:    seriesId,
				Namespace: 5,
			},
			From: now - (86400 * 1000 * 2),
			To:   now + (86400 * 1000 * 1),
		})
		if res.Error != nil {
			t.Error(res.Error)
		}
		var ts uint64
		var val float64
		for ts, val = range res.Results {
			if ts == now {
				break
			}
		}
		if now != ts {
			t.Error("timestamp mismatch", now, ts)
		}
		if writeVal != val {
			t.Error("value mismatch", writeVal, val)
		}
	}
}
