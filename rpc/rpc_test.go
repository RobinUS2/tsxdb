package rpc_test

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"testing"
)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}


func TestNew(t *testing.T) {
	// wait until calls complete
	//wg := sync.WaitGroup{}
	//wg.Add(1)

	// server
	{
		arith := new(Arith)
		rpc.Register(arith)
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", ":1234")
		if e != nil {
			t.Fatal("listen error:", e)
		}
		go http.Serve(l, nil)
	}

	// client
	{
		const serverAddress = "127.0.0.1"
		client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
		if err != nil {
			t.Fatal("dialing:", err)
		}

		// sync
		args := &Args{7,8}
		var reply int
		err = client.Call("Arith.Multiply", args, &reply)
		if err != nil {
			t.Fatal("arith error:", err)
		}
		t.Logf("Arith: %d*%d=%d", args.A, args.B, reply)

		// Asynchronous call
		quotient := new(Quotient)
		divCall := client.Go("Arith.Divide", args, quotient, nil)
		replyCall := <-divCall.Done	// will be equal to divCall
		if replyCall.Error != nil {
			t.Error(replyCall.Error)
		}
		v := replyCall.Reply.(*Quotient)
		t.Logf("Divide %+v", v)
		// check errors, print, etc.
	}
}
