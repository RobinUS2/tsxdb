package client_test

import (
	"github.com/RobinUS2/tsxdb/client"
	"testing"
)

func TestQueryResult_Iterator(t *testing.T) {
	res := client.QueryResult{
		Results: map[uint64]float64{
			1: 2.0,
			4: 5.0,
			2: 3.0,
			0: 3.0,
		},
	}
	n := len(res.Results)
	if n != 4 {
		t.Error(n)
	}

	// verify sorted
	iter := res.Iterator()
	{
		iterN := 0
		for iter.Next() {
			ts, val := iter.Value()
			if iterN == 0 {
				if ts != 0 {
					t.Error(ts)
				}
				if val != 3.0 {
					t.Error()
				}
			} else if iterN == 1 {
				if ts != 1 {
					t.Error(ts)
				}
				if val != 2.0 {
					t.Error()
				}
			} else if iterN == 2 {
				if ts != 2 {
					t.Error(ts)
				}
				if val != 3.0 {
					t.Error()
				}
			} else if iterN == 3 {
				if ts != 4 {
					t.Error(ts)
				}
				if val != 5.0 {
					t.Error()
				}
			}
			iterN++
		}
		if iterN != n {
			t.Error(iterN, n)
		}
	}

	// call again, should be empty now
	if iter.Next() {
		t.Error()
	}

	// reset
	iter.Reset()

	// scan again
	{
		iterN := 0
		for iter.Next() {
			iterN++
		}
		if iterN != 4 {
			t.Error(iterN, 4)
		}
	}
}
