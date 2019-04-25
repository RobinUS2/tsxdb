package backend_test

import (
	"../backend"
	"testing"
)

func TestNewRedisBackend(t *testing.T) {
	b := backend.NewRedisBackend()
	if b == nil {
		t.Error()
	}
}
