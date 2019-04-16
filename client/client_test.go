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
		return
	}

	// new series
	series := c.Series("mySeries")

	// timestamp
	now := c.Now()
	const oneMinute = 60 * 1000

	// write
	{
		result := series.Write(now, 10.1)
		if result.Error != nil {
			t.Error(result.Error)
		}
	}

	// read
	{
		result := series.QueryBuilder().From(now - oneMinute).To(now + oneMinute).Execute()
		if result.Error != nil {
			t.Error(result.Error)
		}
	}
}
