package types

type WriteRequest struct {
	SessionTicket
	Times  []uint64
	Values []float64
	SeriesIdentifier
}

type WriteResponse struct {
	Num   int
	Error *RpcError
}

var EndpointWriter = Endpoint("Writer")
