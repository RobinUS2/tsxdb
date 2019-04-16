package types

type ReadRequest struct {
	From uint64
	To   uint64
	SeriesIdentifier
}

type ReadResponse struct {
	Error *RpcError
}

var EndpointReader = Endpoint("Reader")
