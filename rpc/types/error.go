package types

// go does not (yet) support regular error types over gob

type RpcError string

var RpcErrorNotImplemented RpcError = "not implemented"
var RpcErrorNumTimeValuePairsMisMatch RpcError = "mismatch between number of time&value pairs"

func (err RpcError) String() string {
	return string(err)
}
