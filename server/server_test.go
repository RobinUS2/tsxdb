package server_test

import (
	"../client"
	"../rpc/types"
	"../server"
	"../tools"
	"encoding/base64"
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/rpc"
	"testing"
)

func TestNew(t *testing.T) {
	const token = "verySecure123@#$"
	opts := server.NewOpts()
	opts.AuthToken = token
	s := server.New(opts)
	if s == nil {
		t.Error()
		return
	}
	if err := s.Init(); err != nil {
		t.Error(err)
	}
	if err := s.Start(); err != nil {
		t.Error(err)
	}

	// execute a write
	c, err := rpc.Dial("tcp", opts.ListenHost+fmt.Sprintf(":%d", opts.ListenPort))
	if err != nil {
		t.Fatal("dialing:", err)
	}

	// auth
	var sessionId int
	var sessionSecret []byte
	var authTwoRequest types.AuthRequest
	{
		// 1
		authRequest, _ := client.BasicAuthRequest(opts.OptsConnection)
		var authReply *types.AuthResponse
		err = c.Call(types.EndpointAuth.String()+"."+types.MethodName, authRequest, &authReply)
		if err != nil {
			t.Error("error:", err)
		}
		sessionId = authReply.SessionId
		sessionSecret, _ = base64.StdEncoding.DecodeString(authReply.SessionSecret)
		//log.Printf("%+v", authReply)
	}
	{
		// 2
		authTwoRequest, _ = client.BasicAuthRequest(opts.OptsConnection)
		authTwoRequest.SessionTicket.Id = sessionId
		authTwoRequest.SessionTicket.Nonce = rand.Int()
		authTwoRequest.SessionTicket.Signature = tools.HmacInt(sessionSecret, authTwoRequest.SessionTicket.Nonce)
		var authReply *types.AuthResponse
		err = c.Call(types.EndpointAuth.String()+"."+types.MethodName, authTwoRequest, &authReply)
		if err != nil {
			t.Error("error:", err)
		}
		//log.Printf("%+v", authReply)
	}

	// write
	params := &types.WriteRequest{
		Series: []types.WriteSeriesRequest{{
			Times:  []uint64{1, 2},
			Values: []float64{5.0, 6.0},
		}},
		SessionTicket: authTwoRequest.SessionTicket,
	}
	var reply *types.WriteResponse
	err = c.Call(types.EndpointWriter.String()+"."+types.MethodName, params, &reply)
	if err != nil {
		t.Error("error:", err)
	}
	if reply.Num != 2 {
		t.Error(reply)
	}
	if reply.Error != nil {
		t.Error(reply.Error.String())
	}
	if err := c.Close(); err != nil {
		t.Error(err)
	}
}
