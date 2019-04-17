package rpc_test

import (
	"errors"
	"fmt"
	"net"
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
	const network = "tcp"

	// server
	{
		arith := new(Arith)
		server := rpc.NewServer()
		if err := server.Register(arith); err != nil {
			t.Error(err)
		}
		l, e := net.Listen(network, ":1234")
		if e != nil {
			t.Fatal("listen error:", e)
		}

		go func() {
			for {
				// Get net.TCPConn object
				conn, err := l.Accept()
				if err != nil {
					fmt.Println(err)
					break
				}

				go server.ServeConn(conn)
			}
		}()
	}

	// client
	{
		const serverAddress = "127.0.0.1"
		client, err := rpc.Dial(network, serverAddress+":1234")
		if err != nil {
			t.Fatal("dialing:", err)
		}

		// sync
		args := &Args{7, 8}
		var reply int
		err = client.Call("Arith.Multiply", args, &reply)
		if err != nil {
			t.Fatal("arith error:", err)
		}
		if reply != 56 {
			t.Error(reply)
		}
		//t.Logf("Arith: %d*%d=%d", args.A, args.B, reply)

		// Asynchronous call
		quotient := new(Quotient)
		divCall := client.Go("Arith.Divide", args, quotient, nil)
		replyCall := <-divCall.Done // will be equal to divCall
		if replyCall.Error != nil {
			t.Error(replyCall.Error)
		}
		v := replyCall.Reply.(*Quotient)
		if v.Rem != 7 {
			t.Error(v.Rem)
		}
		//t.Logf("Divide %+v", v)
		// check errors, print, etc.
	}
}
