package rpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net/rpc"
)

type GobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *GobServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *GobServerCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *GobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

func NewGobServerCodec(rwc io.ReadWriteCloser) *GobServerCodec {
	encBuf := bufio.NewWriter(rwc)
	dec := gob.NewDecoder(rwc)
	return &GobServerCodec{
		rwc:    rwc,
		dec:    dec,
		enc:    gob.NewEncoder(rwc),
		encBuf: encBuf,
	}
}
