package types

type SeriesMetadata struct {
	Namespace int
	Name      string
	Tags      []string
	Ttl       uint // relative time in seconds
}

type SeriesCreateMetadata struct {
	SeriesMetadata
	SeriesCreateIdentifier
}

type SeriesMetadataRequest struct {
	// @todo support multiple series at once, will increase performance of first flushes a lot (e.g. batch size of 1000 will have 1000 round trips over TCP, have seen easily 30 seconds for that)
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
