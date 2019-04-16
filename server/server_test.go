package server_test

import (
	"../rpc/types"
	"../server"
	"fmt"
	"net/rpc"
	"testing"
)

func TestNew(t *testing.T) {
	opts := server.NewOpts()
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
	client, err := rpc.Dial("tcp", opts.ListenHost+fmt.Sprintf(":%d", opts.ListenPort))
	if err != nil {
		t.Fatal("dialing:", err)
	}

	// sync
	params := &types.WriteRequest{
		Times:  []uint64{1, 2},
		Values: []float64{5.0, 6.0},
	}
	var reply *types.WriteResponse
	var endpoint = server.NewWriterEndpoint().name().String()
	err = client.Call(endpoint+".Execute", params, &reply)
	if err != nil {
		t.Error("error:", err)
	}
	if reply.Num != 2 {
		t.Error(reply)
	}
	if reply.Error != nil {
		t.Error(reply.Error.String())
	}
	if err := client.Close(); err != nil {
		t.Error(err)
	}
}
