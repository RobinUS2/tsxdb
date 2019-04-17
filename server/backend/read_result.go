package backend

type ReadResult struct {
	Error   error
	Results map[uint64]float64
}
