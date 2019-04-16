package types

type Endpoint string

func (endpoint Endpoint) String() string {
	return string(endpoint)
}

var MethodName string = "Execute"
