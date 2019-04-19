package types

type SeriesMetadataRequest struct {
	Namespace int
	Name      string
	Tags      []string
	SessionTicket
}

type SeriesMetadataResponse struct {
	Id    uint64
	Error *RpcError
}

var EndpointSeriesMetadata = Endpoint("SeriesMetadata")
