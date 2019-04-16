package types

type ReadRequest struct {
	From uint64
	To   uint64
	// @todo series
}

type ReadResponse struct {
	Error *RpcError
}

var EndpointReader = Endpoint("Reader")
