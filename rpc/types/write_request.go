package types

type WriteRequest struct {
	Times  []uint64
	Values []float64
	// @todo series
}

type WriteResponse struct {
	Num   int
	Error error
}
