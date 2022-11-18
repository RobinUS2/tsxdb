package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/tools"
	insecureRand "math/rand"
	"sync"
	"sync/atomic"
)

func init() {
	// init on module load
	registerEndpoint(NewAuthEndpoint())
}

type AuthEndpoint struct {
	server    *Instance
	serverMux sync.RWMutex
}

func (endpoint *AuthEndpoint) getServer() *Instance {
	endpoint.serverMux.RLock()
	s := endpoint.server
	endpoint.serverMux.RUnlock()
	return s
}

func NewAuthEndpoint() *AuthEndpoint {
	return &AuthEndpoint{}
}

func (endpoint *AuthEndpoint) Execute(args *types.AuthRequest, resp *types.AuthResponse) error {
	// deal with panics, else the whole RPC server could crash
	defer func() {
		if r := recover(); r != nil {
			resp.Error = types.WrapErrorPointer(fmt.Errorf("%s", r))
		}
	}()

	nonce, _ := base64.StdEncoding.DecodeString(args.Nonce)
	signature, _ := base64.StdEncoding.DecodeString(args.Signature)

	// signature
	server := endpoint.getServer()
	mac := hmac.New(sha512.New, []byte(server.opts.AuthToken))
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
		server.registerSessionToken(SessionId(resp.SessionId), token)
	} else {
		// stage 2
		if err := server.validateSession(args.SessionTicket); err != nil {
			resp.Error = types.WrapErrorPointer(err)
			return nil
		}

		// auth stats
		atomic.AddUint64(&server.numAuthentications, 1)
	}

	resp.Error = nil
	return nil
}

func (endpoint *AuthEndpoint) register(opts *EndpointOpts) error {
	if err := opts.server.rpc.RegisterName(endpoint.name().String(), endpoint); err != nil {
		return err
	}
	endpoint.serverMux.Lock()
	endpoint.server = opts.server
	endpoint.serverMux.Unlock()
	return nil
}

func (endpoint *AuthEndpoint) name() EndpointName {
	return EndpointName(types.EndpointAuth)
}

func (instance *Instance) validateSession(ticket types.SessionTicket) error {
	if ticket.Id == 0 {
		return errors.New("missing session id")
	}
	if ticket.Nonce == 0 {
		return errors.New("missing session nonce")
	}
	token := instance.getTokenFromSessionId(SessionId(ticket.Id))
	if len(token) != 32 {
		return errors.New("session continuation token not found")
	}
	//log.Printf("token should be %s", base64.StdEncoding.EncodeToString(token))

	// compute of nonce
	expectedSessionSignature := tools.HmacInt(token, ticket.Nonce)
	if expectedSessionSignature != ticket.Signature {
		return types.RpcErrorAuthFailed.Error()
	}

	// track statistics of calls
	atomic.AddUint64(&instance.numCalls, 1)

	return nil
}
