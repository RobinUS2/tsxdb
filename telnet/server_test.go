package telnet_test

import (
	"bytes"
	"github.com/RobinUS2/tsxdb/telnet"
	"io"
	"log"
	"math"
	"testing"
)

type MockReader struct {
	data     chan byte
	shutdown chan bool
}

func (m *MockReader) Read(x []byte) (int, error) {
	xn := len(x)
	cn := len(m.data)
	n := int(math.Min(float64(xn), float64(cn)))
	if n < 1 {
		<-m.shutdown

		// empty
		return 0, io.EOF
	}

	for i := 0; i < n; i++ {
		v := <-m.data
		x[i] = v
	}

	return n, nil
}

func TestInstance_ServeTELNET(t *testing.T) {
	o := telnet.NewOpts()
	o.AuthToken = "verySecure"
	instance := telnet.New(o)
	w := &bytes.Buffer{}
	r := &MockReader{
		data:     make(chan byte, 100),
		shutdown: make(chan bool, 1),
	}
	go func() {
		for {
			r, _ := w.ReadByte()
			if r > 0 {
				log.Printf("read %d %s", r, string(byte(r)))
			}
		}
	}()
	bytesToChan([]byte("bla\r\n"), r.data)
	bytesToChan([]byte("bla bla\r\n"), r.data)
	bytesToChan([]byte("auth wrong\r\n"), r.data)
	bytesToChan([]byte("auth verySecure\r\n"), r.data)
	bytesToChan([]byte("bla\r\n"), r.data)
	bytesToChan([]byte("bla bla\r\n"), r.data)
	bytesToChan([]byte("ECHO bla\r\n"), r.data)
	bytesToChan([]byte("ZADD 1234 10.0\r\n"), r.data)
	instance.Serve(w, r)
}

func bytesToChan(line []byte, c chan byte) {
	for _, b := range line {
		c <- b
	}
}
