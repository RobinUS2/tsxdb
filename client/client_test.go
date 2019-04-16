package client_test

import (
	"../client"
	"testing"
)

func TestNew(t *testing.T) {
	opts := client.NewOpts()
	c := client.New(opts)
	if c == nil {
		t.Error()
	}
}
