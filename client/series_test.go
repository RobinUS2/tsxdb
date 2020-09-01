package client_test

import (
	"github.com/RobinUS2/tsxdb/client"
	"testing"
	"time"
)

func TestNewSeriesTimeout(t *testing.T) {
	c := client.DefaultClient()
	c.SetPreEagerInitFn(getTimeoutFn(client.EagerSeriesInitTimeout + time.Second))
	series := c.Series("test", client.NewSeriesNamespace(1))
	if series.Namespace() != 1 {
		t.Error(series.Namespace())
	}
	time.Sleep(client.EagerSeriesInitTimeout + time.Second)
	if series.InitState() != 1 {
		t.Errorf("expected series to be not created due to timeout %d", series.InitState())
	}
}

func TestNewSeriesPanic(t *testing.T) {
	c := client.DefaultClient()
	c.SetPreEagerInitFn(func(series *client.Series) {
		panic("should error")
	})
	series := c.Series("test", client.NewSeriesNamespace(1))
	if series.Namespace() != 1 {
		t.Error(series.Namespace())
	}
	time.Sleep(time.Second * 1)
	if series.InitState() != 2 {
		t.Errorf("expected series to be not created due to panic %d", series.InitState())
	}
}

func TestNewSeriesSuccess(t *testing.T) {
	c := client.DefaultClient()
	c.SetPreEagerInitFn(func(series *client.Series) {
		panic("should error")
	})
	series := c.Series("test", client.NewSeriesNamespace(1))
	if series.Namespace() != 1 {
		t.Error(series.Namespace())
	}
	time.Sleep(time.Second * 1)
	if series.InitState() != 3 {
		t.Errorf("expected series to be created %d", series.InitState())
	}
}

func getTimeoutFn(duration time.Duration) func(series *client.Series) {
	return func(series *client.Series) {
		time.Sleep(duration)
	}
}
