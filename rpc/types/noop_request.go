package types

type NoOpRequest struct {
	SessionTicket
}

type NoOpResponse struct {
	Error *RpcError
}

var EndpointNoOp = Endpoint("NoOp")
