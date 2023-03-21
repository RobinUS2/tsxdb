package client

import (
	"encoding/base64"
	"fmt"
	tsxdbRpc "github.com/RobinUS2/tsxdb/rpc"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/RobinUS2/tsxdb/tools"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	insecureRand "math/rand"
	"net"
	"net/rpc"
	"strings"
	"sync/atomic"
	"time"
)

// @todo explore something like https://github.com/grpc/grpc OR https://github.com/valyala/gorpc as replacement of regular net/rpc (right now a lot of overhead in network bytes)

func (client *Instance) initConnectionPool() error {
	client.connectionPool = client.NewConnectionPool()
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
		// auth with retries
		err = handleRetry(func() error {
			err := managedConnection.auth(client)
			if err != nil {
				// new connection
				managedConnection.Discard()
				client.connectionPool.Put(managedConnection)
				managedConnection = client.connectionPool.Get().(*ManagedConnection)
				return err
			}
			return nil
		})
		if err != nil {
			// fatal auth error
			return nil, err
		}
	}

	// track timing
	now := nowMs()
	atomic.StoreUint64(&managedConnection.poolGet, now)

	// track potential leakage (e.g not calling *ManagedConnection.Close )
	if client.opts.OptsConnection.Debug {
		const returnTimeout = tsxdbRpc.DefaultTimeout
		time.AfterFunc(returnTimeout, func() {
			returned := atomic.LoadUint64(&managedConnection.poolReturn)
			if returned < now {
				log.Warnf("not returned connection after %s", returnTimeout)
			}
		})
	}

	return managedConnection, nil
}

func (client *Instance) NewClient() (*ManagedConnection, error) {
	// open connection
	address := client.opts.ListenHost + fmt.Sprintf(":%d", client.opts.ListenPort)
	conn, err := net.DialTimeout("tcp", address, client.opts.OptsConnection.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	// codec
	codec := tsxdbRpc.NewGobClientCodec(conn)

	// client
	rpcClient := rpc.NewClientWithCodec(codec)

	return &ManagedConnection{
		service: client,
		client:  rpcClient,
		created: time.Now().Unix(),
	}, nil
}

type ManagedConnection struct {
	service       *Instance
	client        *rpc.Client
	created       int64
	authenticated bool
	sessionId     int
	sessionSecret []byte
	poolGet       uint64
	poolReturn    uint64
	discard       bool // if set to true won't be returned back to the pool
	timesUsed     uint64
}

func (conn *ManagedConnection) DiscardPool() {
	err := conn.Close()
	if err != nil {
		log.Errorf("failed to close connection %s", err)
	}
}

// Discard call this to make sure connection is not reused
func (conn *ManagedConnection) Discard() {
	if conn == nil {
		return
	}
	conn.discard = true
	// this will be discarded in the Close method below
}

func (conn *ManagedConnection) Close() error {
	// track slow usage
	now := nowMs()
	atomic.StoreUint64(&conn.poolReturn, now)
	get := atomic.LoadUint64(&conn.poolGet)
	took := nowMs() - get
	if took > 30*1000 { // @todo configurable
		log.Warnf("SLOW connection usage, taken at %d returned at %d took %d ms", get, now, took) // @todo via logger
	}

	// track max usage per connection
	const maxUsages = 1000
	numUsed := atomic.LoadUint64(&conn.timesUsed)

	// keep alive? only if within expire time and not discard
	if !conn.discard && time.Now().Unix()-conn.created < 60 && numUsed < maxUsages {
		// re-use
		conn.service.connectionPool.Put(conn)
		return nil
	}

	// close
	conn.service.connectionPool.Discard(conn)
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
			errNoSuccess := errors.New("no success")
			if err == nil {
				err = errNoSuccess
			} else {
				err = errors.Wrap(err, errNoSuccess.Error())
			}
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
			return errors.Wrap(err, "failed auth call #1")
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
			return errors.Wrap(err, "failed auth call #2")
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
