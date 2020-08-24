package backend_test

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/server/backend"
	"github.com/go-redis/redis/v7"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const floatTolerance = 0.00001

func TestNewRedisBackendSingleConnection(t *testing.T) {
	opts := &backend.RedisOpts{
		ConnectionDetails: map[backend.Namespace]backend.RedisConnectionDetails{
			backend.RedisDefaultConnectionNamespace: {
				Addr: "127.0.0.1",
				Port: 6379,
				Type: backend.RedisMemory,
			},
		},
	}
	b := backend.NewRedisBackend(opts)
	b.SetReverseApi(b) // we implement this
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
				Type: backend.RedisMemory,
			},
			backend.Namespace(5): {
				Database: 5,
				Type:     backend.RedisMemory,
			},
		},
	}
	b := backend.NewRedisBackend(opts)
	b.SetReverseApi(b) // we implement this
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
	now := uint64(time.Now().Unix() * 1000)
	writeVal := rand.Float64()
	{
		if err := b.Write(backend.ContextWrite{
			Context: backend.Context{
				Namespace: 1,
				Series:    idFirst,
			},
		}, []uint64{now}, []float64{writeVal}); err != nil {
			t.Error(err)
		}
	}

	// simple read
	{
		res := b.Read(backend.ContextRead{
			Context: backend.Context{
				Series:    idFirst,
				Namespace: 1,
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

	// TTL expiry on entire series
	{
		name := fmt.Sprintf("expiry-series-redis-%d", time.Now().UnixNano())
		req := &backend.CreateSeries{
			Series: map[types.SeriesCreateIdentifier]types.SeriesCreateMetadata{
				1: {
					SeriesMetadata: types.SeriesMetadata{
						Name:      name,
						Namespace: 1,
						Ttl:       1, // 1 second
					},
					SeriesCreateIdentifier: 1,
				},
			},
		}

		// first create
		var firstResult types.SeriesMetadataResponse
		{
			resp := b.CreateOrUpdateSeries(req)
			if resp == nil {
				t.Error()
			}
			firstResult = resp.Results[1]
			if firstResult.Id == 0 {
				t.Error(resp.Results)
			}
			if !firstResult.New {
				t.Error("should be new")
			}
		}

		// write
		now := uint64(time.Now().Unix() * 1000)
		writeVal := rand.Float64()
		{
			if err := b.Write(backend.ContextWrite{
				Context: backend.Context{
					Namespace: 1,
					Series:    firstResult.Id,
				},
			}, []uint64{now}, []float64{writeVal}); err != nil {
				t.Error(err)
			}
		}

		// @todo inspect TTL meta

		// read
		{
			res := b.Read(backend.ContextRead{
				Context: backend.Context{
					Series:    firstResult.Id,
					Namespace: 1,
				},
				From: now - 1,
				To:   now + 1,
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

		// wait for expiry
		time.Sleep(2100 * time.Millisecond)

		// read, should be gone
		{
			res := b.Read(backend.ContextRead{
				Context: backend.Context{
					Series:    firstResult.Id,
					Namespace: 1,
				},
				From: now - 1,
				To:   now + 1,
			})
			if res.Error == nil || !strings.Contains(res.Error.Error(), "no data found") {
				t.Error(res.Error)
			}
		}

		// check really removed
		{
			conn := b.GetConnection(1)
			if conn == nil {
				t.Error(conn)
			}
			res := conn.Get(fmt.Sprintf("series_%d_%d_meta", 1, firstResult.Id))
			if res.Err() != redis.Nil || res.Val() != "" {
				t.Error(res.Err(), res.Val())
			}
		}
		// check data removed
		{
			conn := b.GetConnection(1)
			if conn == nil {
				t.Error(conn)
			}
			res := conn.Keys("data_1-2-*")
			for _, key := range res.Val() {
				zrangeRes := conn.ZRange(key, 0, -1)
				if len(zrangeRes.Val()) > 0 {
					ttlRes := conn.PTTL(key)
					t.Errorf("key %s still exists, ttl: %d, data: %v", key, ttlRes.Val().Microseconds(), zrangeRes.Val())
				}
			}
		}

	}
	// end TTL test
}
func CompareFloat(a float64, b float64, tolerance float64, err func()) {
	if math.Abs(a-b) > tolerance {
		err()
	}
}

func TestFloatToString(t *testing.T) {
	tests := map[string]float64{
		"0.1":           0.1,
		"0.123456":      0.123456,
		"0.123457":      0.123456789,      // truncated to 6 digits and rounded
		"123456.123457": 123456.123456789, // truncated to 6 digits and rounded
		"123456.1":      123456.1,
	}
	for expected, float := range tests {
		res := backend.FloatToString(float)
		if res != expected {
			t.Errorf("expected '%s' was '%s'", expected, res)
		}
	}
}
