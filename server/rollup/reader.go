package rollup

import (
	"github.com/Route42/tsxdb/server/backend"
)

type Reader struct {
}

func (reader *Reader) Process(result backend.ReadResult) backend.ReadResult {
	// @Todo real implementation
	return result
}

func NewReader() *Reader {
	return &Reader{}
}
