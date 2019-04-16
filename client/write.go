package client

func (series Series) Write(ts uint64, v float64) (res WriteResult) {
	// @todo implement
	return
}

// @todo batch write

type WriteResult struct {
	Error error
}
