package client

import (
	"errors"
)

var errClientValidationMismatchSent = errors.New("mismatch between expected written values and received")

func (series *Series) Write(ts uint64, v float64) (res WriteResult) {
	// basically a batch with 1 item
	b := series.client.NewBatchWriter()
	if err := b.AddToBatch(series, ts, v); err != nil {
		res.Error = err
		return
	}
	return b.Execute()
}

type WriteResult struct {
	Error        error
	NumPersisted int
}
