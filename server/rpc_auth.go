package server

import (
	"../rpc/types"
	"../tools"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"log"
	insecureRand "math/rand"
)

func init() {
	// init on module load
	registerEndpoint(NewAuthEndpoint())
}

type AuthEndpoint struct {
	server *Instance
}

func NewAuthEndpoint() *AuthEndpoint {
	return &AuthEndpoint{}
}

func (endpoint *AuthEndpoint) Execute(args *types.AuthRequest, resp *types.AuthResponse) error {
	nonce, _ := base64.StdEncoding.DecodeString(args.Nonce)
	signature, _ := base64.StdEncoding.DecodeString(args.Signature)

	// signature
	mac := hmac.New(sha512.New, []byte(endpoint.server.opts.AuthToken))
	mac.Write(nonce)
	expected := mac.Sum(nil)
	if !hmac.Equal(signature, expected) || nonce == nil || len(nonce) < 32 || signature == nil || len(signature) < 1 {
		resp.Error = &types.RpcErrorAuthFailed
		return nil
	}

	// validate stage specific
	if args.SessionTicket.Nonce == 0 {
		// stage 1
		var token = make([]byte, 32)
		if _, err := rand.Read(token); err != nil {
			resp.Error = types.WrapErrorPointer(errors.New("entropy error"))
		}
		if len(token) != 32 {
			panic("token length")
		}
		// session ID (non-zero)
		for {
			resp.SessionId = insecureRand.Int()
			if resp.SessionId != 0 {
				break
			}
		}
		resp.SessionSecret = base64.StdEncoding.EncodeToString(token)

		// store in server
		// @todo token expiry
		endpoint.server.sessionTokensMux.Lock()
		endpoint.server.sessionTokens[resp.SessionId] = token
		endpoint.server.sessionTokensMux.Unlock()
	} else {
		// stage 2
		log.Printf("stage 2 %+v", args)
		if args.SessionTicket.Id == 0 {
			resp.Error = types.WrapErrorPointer(errors.New("missing session id"))
			return nil
		}
		if args.SessionTicket.Nonce == 0 {
			resp.Error = types.WrapErrorPointer(errors.New("missing session nonce"))
			return nil
		}
		endpoint.server.sessionTokensMux.RLock()
		token := endpoint.server.sessionTokens[args.SessionTicket.Id]
		endpoint.server.sessionTokensMux.RUnlock()
		if len(token) != 32 {
			resp.Error = types.WrapErrorStringPointer("session continuation token not found")
			return nil
		}
		log.Printf("token should be %s", base64.StdEncoding.EncodeToString(token))

		// compute of nonce
		expectedSessionSignature := tools.HmacInt(token, args.SessionTicket.Nonce)
		if expectedSessionSignature != args.SessionTicket.Signature {
			resp.Error = &types.RpcErrorAuthFailed
			return nil
		}
	}

	resp.Error = nil
	return nil
}

func (endpoint *AuthEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.server = opts.server
	return nil
}

func (endpoint *AuthEndpoint) name() EndpointName {
	return EndpointName(types.EndpointAuth)
}
