package client_test

import (
	"github.com/Route42/tsxdb/client"
	"testing"
)

func TestNewSeriesNamespace(t *testing.T) {
	c := client.DefaultClient()
	series := c.Series("test", client.NewSeriesNamespace(1))
	if series.Namespace() != 1 {
		t.Error(series.Namespace())
	}
}

func TestNewSeriesNoNamespace(t *testing.T) {
	c := client.DefaultClient()
	series := c.Series("test")
	if series.Namespace() != 0 {
		t.Error("expected default 0", series.Namespace())
	}
}
