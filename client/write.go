package client

import (
	"../rpc/types"
	"log"
)

func (series Series) Write(ts uint64, v float64) (res WriteResult) {
	// @todo implement
	writeRequest := types.WriteRequest{}
	log.Printf("%+v", writeRequest)
	return
}

// @todo batch write

type WriteResult struct {
	Error error
}
