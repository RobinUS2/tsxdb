package client_test

import (
	"github.com/RobinUS2/tsxdb/client"
	"testing"
)

func TestNewSeriesPool(t *testing.T) {
	c := client.New(client.NewOpts())
	pool := c.SeriesPool()
	if pool.Count() != 0 {
		t.Error(pool.Count())
	}
	if pool.Hits() != 0 {
		t.Error(pool.Hits())
	}

	// first fetch
	{
		s := client.NewSeries("test", c)
		if s == nil {
			t.Error()
		}
		if pool.Count() != 1 {
			t.Error(pool.Count())
		}
		if pool.Hits() != 0 {
			t.Error(pool.Hits())
		}
	}

	// fetch again
	{
		s := client.NewSeries("test", c)
		if s == nil {
			t.Error()
		}
		if pool.Count() != 1 {
			t.Error(pool.Count())
		}
		if pool.Hits() != 1 {
			t.Error(pool.Hits())
		}
	}

	// first fetch new item
	{
		s := client.NewSeries("test2", c)
		if s == nil {
			t.Error()
		}
		if pool.Count() != 2 {
			t.Error(pool.Count())
		}
		if pool.Hits() != 1 {
			t.Error(pool.Hits())
		}
	}

	// second fetch second item
	{
		s := client.NewSeries("test2", c)
		if s == nil {
			t.Error()
		}
		if pool.Count() != 2 {
			t.Error(pool.Count())
		}
		if pool.Hits() != 2 {
			t.Error(pool.Hits())
		}
	}
}
