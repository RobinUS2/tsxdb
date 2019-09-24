package types

import "errors"

// go does not (yet) support regular error types over gob

type RpcError string

var RpcErrorNotImplemented RpcError = "not implemented"
var RpcErrorAuthFailed RpcError = "not authenticated"
var RpcErrorNumTimeValuePairsMisMatch RpcError = "mismatch between number of time&value pairs"
var RpcErrorNoValues RpcError = "no values"
var RpcErrorMissingSeriesId RpcError = "missing series id"
var RpcErrorBackendStrategyNotFound RpcError = "no backend strategy found"
var RpcErrorBackendMetadataNotFound RpcError = "missing metadata"

func (err RpcError) String() string {
	return string(err)
}

func (err RpcError) Error() error {
	return errors.New(err.String())
}

func WrapErrorPointer(err error) *RpcError {
	return WrapErrorStringPointer(err.Error())
}

func WrapErrorStringPointer(err string) *RpcError {
	e := RpcError(err)
	return &e
}
