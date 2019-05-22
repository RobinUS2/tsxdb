package client

import (
	"encoding/base64"
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/tools"
	"github.com/pkg/errors"
	insecureRand "math/rand"
	"net/rpc"
	"strings"
	"sync"
	"time"
)

func (client *Instance) initConnectionPool() error {
	client.connectionPool = &sync.Pool{
		New: func() interface{} {
			if client.closing {
				return nil
			}
			c, err := client.NewConnection()
			if err != nil {
				// is dealt with in (client *Instance) GetConnection() (*ManagedConnection, error)
				panic(errors.Wrap(err, "failed to init new connection"))
			}
			return c
		},
	}
	return nil
}

func (client *Instance) GetConnection() (managedConnection *ManagedConnection, err error) {
	defer func() {
		if r := recover(); r != nil {
			managedConnection = nil
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	// get connection, this may panic
	conn := client.connectionPool.Get()
	if conn == nil {
		if client.closing {
			return nil, nil
		}
		return nil, errors.New("failed to obtain connection")
	}

	// correct type
	managedConnection = conn.(*ManagedConnection)

	// auth
	if !managedConnection.authenticated {
		if err := managedConnection.auth(client); err != nil {
			return nil, err
		}
	}
	return managedConnection, nil
}

func (client *Instance) NewConnection() (*ManagedConnection, error) {
	conn, err := rpc.Dial("tcp", client.opts.ListenHost+fmt.Sprintf(":%d", client.opts.ListenPort))
	if err != nil {
		return nil, err
	}
	return &ManagedConnection{
		client:  conn,
		pool:    client.connectionPool,
		created: time.Now().Unix(),
	}, nil
}

type ManagedConnection struct {
	client        *rpc.Client
	pool          *sync.Pool
	created       int64
	authenticated bool
	sessionId     int
	sessionSecret []byte
}

func (conn *ManagedConnection) Close() error {
	// keep alive?
	if time.Now().Unix()-conn.created >= 60 {
		// re-use
		conn.pool.Put(conn)
		return nil
	}

	// close
	if err := conn.client.Close(); err != nil {
		return err
	}
	return nil
}

func (conn *ManagedConnection) executeAuthRequest(request types.AuthRequest) (response *types.AuthResponse, err error) {
	success := false
	defer func() {
		// close the real underlying RPC connection
		if !success {
			_ = conn.client.Close()
			err = errors.New("no success")
		}
	}()

	// execute
	if err := conn.client.Call(types.EndpointAuth.String()+"."+types.MethodName, request, &response); err != nil {
		return nil, err
	}

	// pass back
	if response.Error != nil {
		return nil, response.Error.Error()
	}

	success = true
	// good
	return response, nil
}

func (conn *ManagedConnection) auth(client *Instance) error {
	// first stage
	var sessionId int
	var sessionSecret []byte
	{
		// phase 1 initial payload
		request, err := tools.BasicAuthRequest(client.opts.OptsConnection)
		if err != nil {
			return err
		}

		// execute phase 1
		resp, err := conn.executeAuthRequest(request)
		if err != nil {
			return err
		}

		// validate and parse session data
		if resp.SessionId == 0 {
			return errors.New("missing session id")
		}
		sessionId = resp.SessionId
		if len(strings.TrimSpace(resp.SessionSecret)) < 1 {
			return errors.New("missing session secret")
		}
		sessionSecret, err = base64.StdEncoding.DecodeString(resp.SessionSecret)
		if err != nil {
			return err
		}
		//log.Printf("resp stage 1 %+v", resp)
	}

	// second stage
	{
		request, err := tools.BasicAuthRequest(client.opts.OptsConnection)
		if err != nil {
			return err
		}
		request.SessionTicket.Id = sessionId
		request.SessionTicket.Nonce = insecureRand.Int()

		// signature of nonce
		request.SessionTicket.Signature = tools.HmacInt(sessionSecret, request.SessionTicket.Nonce)

		if _, err := conn.executeAuthRequest(request); err != nil {
			return err
		}
		// store for next requests
		conn.sessionId = sessionId
		conn.sessionSecret = sessionSecret
	}
	//log.Println("auth complete")

	return nil
}

func (conn *ManagedConnection) getSessionTicket() types.SessionTicket {
	nonce := insecureRand.Int()
	return types.SessionTicket{
		Id:        conn.sessionId,
		Nonce:     nonce,
		Signature: tools.HmacInt(conn.sessionSecret, nonce),
	}
}

func (client *Instance) Close() {
	client.closing = true
	for {
		conn, _ := client.GetConnection()
		if conn != nil {
			_ = conn.client.Close()
		} else {
			break
		}
	}
}
