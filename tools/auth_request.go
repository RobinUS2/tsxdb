package tools

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/RobinUS2/tsxdb/rpc"
	"github.com/RobinUS2/tsxdb/rpc/types"
)

func BasicAuthRequest(opts rpc.OptsConnection) (request types.AuthRequest, err error) {
	// random data
	var nonce = make([]byte, 32)
	_, err = rand.Read(nonce)
	if err != nil {
		// missing entropy, risky
		return
	}

	// signature
	signature, _ := Hmac([]byte(opts.AuthToken), nonce)

	// request (single)
	request = types.AuthRequest{
		Nonce:     base64.StdEncoding.EncodeToString(nonce),
		Signature: base64.StdEncoding.EncodeToString(signature),
	}
	return
}
