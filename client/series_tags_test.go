package client_test

import (
	"../client"
	"testing"
)

func TestNewSeriesTags(t *testing.T) {
	c := client.DefaultClient()
	series := c.Series("test", client.NewSeriesTags("apple", "banana"))
	tags := series.Tags()
	if len(tags) != 2 {
		t.Error(series)
	}
	if tags[0] != "apple" {
		t.Error()
	}
	if tags[1] != "banana" {
		t.Error()
	}
}

func TestNewSeriesNoTags(t *testing.T) {
	c := client.DefaultClient()
	series := c.Series("test")
	if series.Tags() != nil {
		t.Error("expected nil")
	}
}
