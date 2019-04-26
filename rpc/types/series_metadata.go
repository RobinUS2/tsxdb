package types

type SeriesMetadata struct {
	Namespace int
	Name      string
	Tags      []string
}

type SeriesCreateMetadata struct {
	SeriesMetadata
	SeriesCreateIdentifier
}

type SeriesMetadataRequest struct {
	SeriesCreateMetadata
	SessionTicket
}

type SeriesMetadataResponse struct {
	Id uint64
	SeriesCreateIdentifier
	Error *RpcError
	New   bool
}

type SeriesCreateIdentifier uint64 // xxhash64 of uuid bytes

var EndpointSeriesMetadata = Endpoint("SeriesCreateMetadata")
