package telnet

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/reiver/go-oi"
	tel "github.com/reiver/go-telnet" // weird things happen if package with same name is imported as the package/module it's in unless aliased
	"io"
	"log"
	"strings"
)

type Instance struct {
	opts *Opts
}

func (instance *Instance) Listen() error {
	if len(strings.TrimSpace(instance.opts.AuthToken)) < 1 {
		return errors.New("missing auth token")
	}
	listenStr := fmt.Sprintf("%s:%d", instance.opts.Host, instance.opts.Port)
	log.Printf("telnet listening at %s", listenStr)
	err := tel.ListenAndServe(listenStr, instance)
	if nil != err {
		return err
	}
	return nil
}

func (instance *Instance) ServeTELNET(ctx tel.Context, w tel.Writer, r tel.Reader) {
	lineBuffer := &bytes.Buffer{}
	newBytes := make(chan []byte, 1)
	lines := make(chan string, 1)
	session := NewSession(instance)
	session.SetWriter(w)

	// into line buffer
	const newline = '\n'
	go func() {
		for {
			b := <-newBytes
			if nWritten, err := oi.LongWrite(lineBuffer, b); err != nil || int64(len(b)) != nWritten {
				panic("failed to write")
			}
			if !strings.Contains(string(b), string(newline)) {
				continue
			}
			str, err := lineBuffer.ReadString(newline)
			if err != nil && err != io.EOF {
				panic(err)
			}
			line := strings.TrimSpace(str)
			if len(line) < 1 {
				// empty / whitespaces
				continue
			}
			lines <- line
		}
	}()

	// handle lines
	go func() {
		for {
			line := <-lines
			if err := session.Handle(InputLine(line)); err != nil {
				// pass error
				if err := session.Write(err.Error()); err != nil {
					panic(err)
				}
			}
		}
	}()

	// read from socket, this must be last while loop since it will block the other go-routines above from exiting the function
	// on socket close the rest will stop as well
	for {
		var readBuffer = make([]byte, 1)
		n, err := r.Read(readBuffer)
		if n > 0 {
			readBytes := readBuffer[:n]

			// write to line buffer
			newBytes <- readBytes
		}

		if nil != err {
			break
		}
	}
}

func New(opts *Opts) *Instance {
	return &Instance{
		opts: opts,
	}
}
