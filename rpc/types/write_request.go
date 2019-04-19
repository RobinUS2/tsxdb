package types

type WriteRequest struct {
	SessionTicket
	Series []WriteSeriesRequest
}

type WriteSeriesRequest struct {
	SeriesIdentifier
	Times  []uint64
	Values []float64
}

type WriteResponse struct {
	Num   int
	Error *RpcError
}

var EndpointWriter = Endpoint("Writer")
