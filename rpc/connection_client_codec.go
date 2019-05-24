package rpc

import (
	"bufio"
	"encoding/gob"
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
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *GobClientCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c *GobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobClientCodec) Close() error {
	return c.rwc.Close()
}

func NewGobClientCodec(rwc io.ReadWriteCloser,
	dec *gob.Decoder,
	enc *gob.Encoder,
	encBuf *bufio.Writer) *GobClientCodec {
	return &GobClientCodec{
		rwc:    rwc,
		dec:    dec,
		enc:    enc,
		encBuf: encBuf,
	}
}
