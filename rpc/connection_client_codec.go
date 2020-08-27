package rpc

import (
	"bufio"
	"encoding/gob"
	"github.com/pkg/errors"
	"io"
	"net/rpc"
)

type GobClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *GobClientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		err = errors.Wrap(err, "write request failed to encode request")
		return
	}
	if err = c.enc.Encode(body); err != nil {
		err = errors.Wrap(err, "write request failed to encode body")
		return
	}
	return c.encBuf.Flush()
}

func (c *GobClientCodec) ReadResponseHeader(r *rpc.Response) error {
	err := c.dec.Decode(r)
	if err != nil {
		return errors.Wrap(err, "read response header")
	}
	return nil
}

func (c *GobClientCodec) ReadResponseBody(body interface{}) error {
	err := c.dec.Decode(body)
	if err != nil {
		return errors.Wrap(err, "read response body")
	}
	return nil
}

func (c *GobClientCodec) Close() error {
	err := c.rwc.Close()
	if err != nil {
		return errors.Wrap(err, "close connection codec")
	}
	return nil
}

func NewGobClientCodec(rwc io.ReadWriteCloser) *GobClientCodec {
	encBuf := bufio.NewWriter(rwc)
	dec := gob.NewDecoder(rwc)
	enc := gob.NewEncoder(encBuf)

	return &GobClientCodec{
		rwc:    rwc,
		dec:    dec,
		enc:    enc,
		encBuf: encBuf,
	}
}
