package backend_test

import (
	"../backend"
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"math"
	"math/rand"
	"testing"
	"time"
)

const floatTolerance = 0.00001

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

	// create
	var idFirst uint64
	name := fmt.Sprintf("series-redis-%d", time.Now().UnixNano())
	{
		createIdentifier := types.SeriesCreateIdentifier(1234)
		res := b.CreateOrUpdateSeries(&backend.CreateSeries{
			Series: map[types.SeriesCreateIdentifier]types.SeriesCreateMetadata{
				createIdentifier: {
					SeriesMetadata: types.SeriesMetadata{
						Namespace: 1,
						Name:      name,
						Tags:      []string{"a", "b"},
					},
					SeriesCreateIdentifier: createIdentifier,
				},
			},
		})
		if res == nil {
			t.Error()
		}
		if res.Error != nil {
			t.Error(res.Error)
		}
		result := res.Results[createIdentifier]
		if result.Id < 1 {
			t.Error("missing id")
		}
		idFirst = result.Id
		if !result.New {
			t.Error("should be new")
		}
	}

	// create again
	{
		createIdentifier := types.SeriesCreateIdentifier(2345)
		res := b.CreateOrUpdateSeries(&backend.CreateSeries{
			Series: map[types.SeriesCreateIdentifier]types.SeriesCreateMetadata{
				createIdentifier: {
					SeriesMetadata: types.SeriesMetadata{
						Namespace: 1,
						Name:      name,
						Tags:      []string{"a", "b"},
					},
					SeriesCreateIdentifier: createIdentifier,
				},
			},
		})
		if res == nil {
			t.Error()
		}
		if res.Error != nil {
			t.Error(res.Error)
		}
		result := res.Results[createIdentifier]
		if result.Id != idFirst {
			t.Error(result.Id, idFirst)
		}
		if result.New {
			t.Error("should not be new")
		}
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
		CompareFloat(writeVal, val, floatTolerance, func() {
			t.Error("value mismatch", writeVal, val)
		})
	}

	// search name
	{
		res := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  1,
				Comparator: backend.SearchSeriesComparatorEquals,
				Name:       name,
			},
		})
		if res == nil {
			t.Error(res)
		}
		if res.Error != nil {
			t.Error(res.Error)
		}
		if res.Series == nil {
			t.Error()
		}
		if len(res.Series) != 1 {
			t.Error(res.Series)
		}
		first := res.Series[0]
		if first.Id != idFirst {
			t.Error(first.Id, idFirst)
		}
		if first.Namespace != 1 {
			t.Error(first.Namespace)
		}
	}

	// search name wrong namespace
	{
		res := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  2,
				Comparator: backend.SearchSeriesComparatorEquals,
				Name:       name,
			},
		})
		if res == nil {
			t.Error(res)
		}
		if res.Error != nil {
			t.Error(res.Error)
		}
		if res.Series != nil {
			t.Error("should be nil")
		}
	}

	// search name non existing
	{
		res := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  1,
				Comparator: backend.SearchSeriesComparatorEquals,
				Name:       name + "notExisting",
			},
		})
		if res == nil {
			t.Error(res)
		}
		if res.Error != nil {
			t.Error(res.Error)
		}
		if res.Series != nil {
			t.Error("should be nil")
		}
	}

	// delete
	{
		res := b.DeleteSeries(&backend.DeleteSeries{
			Series: []types.SeriesIdentifier{
				{
					Namespace: 1,
					Id:        idFirst,
				},
			},
		})
		if res == nil {
			t.Error(res)
		}
		if res.Error != nil {
			t.Error(res.Error)
		}
	}

	// search name
	{
		res := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  1,
				Comparator: backend.SearchSeriesComparatorEquals,
				Name:       name,
			},
		})
		if res == nil {
			t.Error(res)
		}
		if res.Error != nil {
			t.Error(res.Error)
		}
		if res.Series != nil {
			t.Error("should be nil")
		}
	}
}
func CompareFloat(a float64, b float64, tolerance float64, err func()) {
	if math.Abs(a-b) > tolerance {
		err()
	}
}
