package server_test

import (
	"../rpc/types"
	"../server"
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

	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Fatal("dialing:", err)
	}

	// sync
	params := &types.WriteRequest{
		Times:  []uint64{1, 2},
		Values: []float64{5.0, 6.0},
	}
	var reply *types.WriteResponse
	err = client.Call("WriterEndpoint.Write", params, &reply)
	if err != nil {
		t.Fatal("arith error:", err)
	}
	if reply.Num != 2 {
		t.Error(reply)
	}
	if reply.Error != nil {
		t.Error(reply.Error.String())
	}
}
