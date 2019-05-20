package backend_test

import (
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/server/backend"
	"math/rand"
	"testing"
	"time"
)

func TestMemoryBackend(t *testing.T) {
	b := backend.NewMemoryBackend()

	// create
	req := &backend.CreateSeries{
		Series: map[types.SeriesCreateIdentifier]types.SeriesCreateMetadata{
			12345: {
				SeriesMetadata: types.SeriesMetadata{
					Tags:      []string{"one", "two"},
					Name:      "banana",
					Namespace: 1,
				},
				SeriesCreateIdentifier: 12345,
			},
		},
	}

	// first create
	{
		resp := b.CreateOrUpdateSeries(req)
		if resp == nil {
			t.Error()
		}
		firstResult := resp.Results[12345]
		if firstResult.Id == 0 {
			t.Error(resp.Results)
		}
		if !firstResult.New {
			t.Error("should be new")
		}
		if firstResult.SeriesCreateIdentifier != 12345 {
			t.Error("should have same reference number")
		}
	}

	// second create (actually not a create :-)
	{
		resp := b.CreateOrUpdateSeries(req)
		if resp == nil {
			t.Error()
		}
		firstResult := resp.Results[12345]
		if firstResult.Id != 1 {
			t.Error(resp.Results)
		}
		if firstResult.New {
			t.Error("should not be new")
		}
		if firstResult.SeriesCreateIdentifier != 12345 {
			t.Error("should have same reference number")
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

	// search by name
	{
		resp := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  1,
				Name:       "banana",
				Comparator: backend.SearchSeriesComparatorEquals,
			},
		})
		if resp == nil {
			t.Error()
		}
		if resp.Error != nil {
			t.Error(resp.Error)
		}
		if resp.Series == nil {
			t.Error()
		}
		if len(resp.Series) != 1 {
			t.Error(len(resp.Series))
		}
		firstResult := resp.Series[0]
		if firstResult.Namespace != 1 {
			t.Error(firstResult)
		}
		if firstResult.Id != 1 {
			t.Error(firstResult)
		}
	}

	// search by name (no match)
	{
		resp := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  1,
				Name:       "notBanana",
				Comparator: backend.SearchSeriesComparatorEquals,
			},
		})
		if resp == nil {
			t.Error()
		}
		if resp.Error != nil {
			t.Error(resp.Error)
		}
		if resp.Series != nil {
			t.Error("should be empty")
		}
	}

	// search by name (wrong namespace)
	{
		resp := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  2,
				Name:       "banana",
				Comparator: backend.SearchSeriesComparatorEquals,
			},
		})
		if resp == nil {
			t.Error()
		}
		if resp.Error != nil {
			t.Error(resp.Error)
		}
		if resp.Series != nil {
			t.Error("should be empty")
		}
	}

	// delete wrong namespace
	{
		resp := b.DeleteSeries(&backend.DeleteSeries{
			Series: []types.SeriesIdentifier{
				{
					Id:        1,
					Namespace: 2,
				},
			},
		})
		if resp == nil {
			t.Error()
		}
		if resp.Error == nil {
			t.Error("expect error")
		}
	}

	// delete non existing
	{
		resp := b.DeleteSeries(&backend.DeleteSeries{
			Series: []types.SeriesIdentifier{
				{
					Id:        15,
					Namespace: 1,
				},
			},
		})
		if resp == nil {
			t.Error()
		}
		if resp.Error == nil {
			t.Error("expect error")
		}
	}

	// delete
	{
		resp := b.DeleteSeries(&backend.DeleteSeries{
			Series: []types.SeriesIdentifier{
				{
					Id:        1,
					Namespace: 1,
				},
			},
		})
		if resp == nil {
			t.Error()
		}
		if resp.Error != nil {
			t.Error(resp.Error)
		}
	}

	// search by name (after deletion)
	{
		resp := b.SearchSeries(&backend.SearchSeries{
			SearchSeriesElement: backend.SearchSeriesElement{
				Namespace:  1,
				Name:       "banana",
				Comparator: backend.SearchSeriesComparatorEquals,
			},
		})
		if resp == nil {
			t.Error()
		}
		if resp.Error != nil {
			t.Error(resp.Error)
		}
		if resp.Series != nil {
			t.Error("should be empty")
		}
	}
}
