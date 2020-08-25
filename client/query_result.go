package client

import "sort"

type QueryResult struct {
	Series  *Series
	Error   error
	Results map[uint64]float64 // in random order due to Go map implementation, if you need sorted results call QueryResult.Iterator()
}

func (res QueryResult) Iterator() *QueryResultIterator {
	iter := &QueryResultIterator{
		results: &res,
		size:    len(res.Results),
	}
	iter.Reset()

	// sort
	sortedKeys := make([]uint64, iter.size)
	idx := 0
	for k := range iter.results.Results {
		sortedKeys[idx] = k
		idx++
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })
	iter.dataKeys = sortedKeys

	return iter
}

type QueryResultIterator struct {
	results  *QueryResult
	current  int
	size     int
	dataKeys []uint64
}

func (iter *QueryResultIterator) Reset() {
	iter.current = -1 // start before first, since you do for it.Next() { it.Value() }
}

func (iter *QueryResultIterator) Next() bool {
	iter.current++
	return iter.current < iter.size
}

func (iter *QueryResultIterator) Value() (uint64, float64) {
	timestamp := iter.dataKeys[iter.current]
	value := iter.results.Results[timestamp]
	return timestamp, value
}
