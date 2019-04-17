package types

type NoOpRequest struct {
}

type NoOpResponse struct {
	Error *RpcError
}

var EndpointNoOp = Endpoint("NoOp")
