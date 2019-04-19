package types

type ReadRequest struct {
	SessionTicket
	Queries []ReadSeriesRequest
}

type ReadSeriesRequest struct {
	From uint64
	To   uint64
	SeriesIdentifier
}

type ReadResponse struct {
	Error   *RpcError
	Results map[uint64]map[uint64]float64 // map series id => timestamp => value
}

var EndpointReader = Endpoint("Reader")
