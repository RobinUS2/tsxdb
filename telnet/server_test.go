package telnet_test

import (
	"errors"
	"fmt"
	"github.com/RobinUS2/tsxdb/server"
	"github.com/RobinUS2/tsxdb/telnet"
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

type MockWriter struct {
	output chan string
}

func (m *MockWriter) Write(b []byte) (int, error) {
	m.output <- string(b)
	return len(b), nil
}

type MockReader struct {
	data     chan byte
	shutdown chan bool
}

func (m *MockReader) Read(x []byte) (int, error) {
	xn := len(x)
	cn := len(m.data)
	n := int(math.Min(float64(xn), float64(cn)))
	if n < 1 {
		select {
		case <-m.shutdown:
			// done
			return 0, io.EOF
		case <-time.After(50 * time.Millisecond):
			// try again
			return m.Read(x)
		}
	}

	for i := 0; i < n; i++ {
		v := <-m.data
		x[i] = v
	}

	return n, nil
}

func TestInstance_ServeTELNET(t *testing.T) {
	// start server
	serverOpts := server.NewOpts()
	serverOpts.AuthToken = "verySecure"
	s := server.New(serverOpts)
	if err := s.Init(); err != nil {
		t.Error(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal("server could not be started", err)
	}
	defer func() {
		_ = s.Shutdown()
	}()

	// telnet proxy
	o := telnet.NewOpts()
	o.AuthToken = serverOpts.AuthToken
	o.ServerPort = serverOpts.ListenPort
	o.ServerHost = serverOpts.ListenHost
	instance := telnet.New(o)
	w := &MockWriter{
		output: make(chan string, 1),
	}
	r := &MockReader{
		data:     make(chan byte, 100),
		shutdown: make(chan bool, 1),
	}
	mustBeError := func(s string) error {
		if !strings.HasPrefix(s, "-ERR") {
			return errors.New("should be error")
		}
		return nil
	}
	mustBeOk := func(s string) error {
		if strings.TrimSpace(s) != "+OK" {
			return errors.New("should be +OK")
		}
		return nil
	}
	mustBeIntOne := func(s string) error {
		if strings.TrimSpace(s) != ":1" {
			return errors.New("should be :1")
		}
		return nil
	}
	tests := []*test{
		{
			cmd:          "bla",
			validationFn: mustBeError,
		},
		{
			cmd:          "bla bla",
			validationFn: mustBeError,
		},
		{
			cmd:          "auth wrong",
			validationFn: mustBeError,
		},
		{
			cmd:          "auth verySecure",
			validationFn: mustBeOk,
		},
		{
			cmd:          "bla",
			validationFn: mustBeError,
		},
		{
			cmd:          "bla bla",
			validationFn: mustBeError,
		},
		{
			cmd: "ECHO bla",
			validationFn: func(s string) error {
				if strings.TrimSpace(s) != "-ERR bla" {
					return errors.New("wrong")
				}
				return nil
			},
		},
		{
			cmd:          "ZADD testSeries 1558110305 10.0",
			validationFn: mustBeIntOne,
		},
		{
			cmd: "ZRANGEBYSCORE testSeries 1558110304 1558110306",
			validationFn: func(s string) error {
				const expect = "*1\r\n$2\r\n10"
				if strings.TrimSpace(s) != expect {
					return fmt.Errorf("should be %s", expect)
				}
				return nil
			},
		},
		{
			cmd: "ZRANGEBYSCORE testSeries 1558110304 1558110306 withscores",
			validationFn: func(s string) error {
				const expect = "*2\r\n$2\r\n10\r\n$10\r\n1558110305"
				if strings.TrimSpace(s) != expect {
					return fmt.Errorf("should be %s", expect)
				}
				return nil
			},
		},
		{
			cmd: "ZRANGEBYSCORE testSeries 1558110304 1558110306 WITHSCORES",
			validationFn: func(s string) error {
				const expect = "*2\r\n$2\r\n10\r\n$10\r\n1558110305"
				if strings.TrimSpace(s) != expect {
					return fmt.Errorf("should be %s", expect)
				}
				return nil
			},
		},
		// add second value
		{
			cmd:          "ZADD testSeries 1558110306 10.1",
			validationFn: mustBeIntOne,
		},
		// add third value
		{
			cmd:          "ZADD testSeries 1558110307 110.5",
			validationFn: mustBeIntOne,
		},
		// get all three without scores
		{
			cmd: "ZRANGEBYSCORE testSeries 1558110304 1558110308",
			validationFn: func(s string) error {
				const expect = "*3\r\n$2\r\n10\r\n$4\r\n10.1\r\n$5\r\n110.5"
				if strings.TrimSpace(s) != expect {
					return fmt.Errorf("should be %s", expect)
				}
				return nil
			},
		},
		// get all three -inf +inf
		{
			cmd: "ZRANGEBYSCORE testSeries -inf +inf",
			validationFn: func(s string) error {
				const expect = "*3\r\n$2\r\n10\r\n$4\r\n10.1\r\n$5\r\n110.5"
				if strings.TrimSpace(s) != expect {
					return fmt.Errorf("should be %s", expect)
				}
				return nil
			},
		},
	}
	testI := 0
	var currentTest *test
	nextTest := func() {
		if testI > len(tests)-1 {
			// done
			r.shutdown <- true
			return
		}
		currentTest = tests[testI]
		testI += 1 // increment to next test
		bytesToChan([]byte(currentTest.cmd+"\r\n"), r.data)
	}
	nextTest() // start it off

	go func() {
		for {
			outLine := <-w.output
			if err := currentTest.validate(outLine); err != nil {
				t.Error(currentTest, outLine, err)
			}
			nextTest()
		}
	}()
	instance.Serve(w, r)
}

func bytesToChan(line []byte, c chan byte) {
	for _, b := range line {
		c <- b
	}
}

type test struct {
	cmd          string
	validationFn func(s string) error
}

func (test *test) validate(s string) error {
	return test.validationFn(s)
}
