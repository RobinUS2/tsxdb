package types

type ReadRequest struct {
	From uint64
	To   uint64
	SeriesIdentifier
}

type ReadResponse struct {
	Error   *RpcError
	Results map[uint64]float64
}

var EndpointReader = Endpoint("Reader")
