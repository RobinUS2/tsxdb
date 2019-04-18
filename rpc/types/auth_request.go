package types

type AuthRequest struct {
	SessionTicket // this will be empty 1st request, 2nd request of the auth handshake it will validate
	Nonce         string
	Signature     string
}

type AuthResponse struct {
	Error         *RpcError
	SessionId     int
	SessionSecret string
}

var EndpointAuth = Endpoint("Auth")
