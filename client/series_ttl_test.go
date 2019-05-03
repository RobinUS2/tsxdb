package client_test

import (
	"github.com/RobinUS2/tsxdb/client"
	"testing"
)

func TestNewSeriesWithoutTTL(t *testing.T) {
	c := client.DefaultClient()
	series := c.Series("test")
	if series.TTL() != 0 {
		t.Error(series.TTL())
	}
}

func TestNewSeriesWithTTL(t *testing.T) {
	c := client.DefaultClient()
	series := c.Series("test", client.NewSeriesTTL(86400))
	if series.TTL() != 86400 {
		t.Error(series.TTL())
	}
}
