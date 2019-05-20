package telnet

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/reiver/go-oi"
	tel "github.com/reiver/go-telnet" // weird things happen if package with same name is imported as the package/module it's in unless aliased
	"io"
	"log"
	"strconv"
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

func (instance *Instance) Serve(w tel.Writer, r tel.Reader) {
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
		const recentLineBufferSize = 5
		var recentLines = make([]string, recentLineBufferSize)
		var recentLineIdx = 0
		var detectingRedis = false
		var firstLine = true
		var lineTokens = make([]string, 0)
		var expectingTuples = uint(0)
		for {
			line := <-lines
			recentLineIdx = (recentLineIdx + 1) % recentLineBufferSize
			recentLines[recentLineIdx] = line

			// detect redis, essentially it starts with *1 $7 COMMAND
			if firstLine {
				firstLine = false
				if line == "*1" {
					// smells like redis
					detectingRedis = true
					continue
				}
			}
			if detectingRedis && line == "$7" {
				// ignore the length of the "COMMAND" word
				continue
			}
			if detectingRedis && line == "COMMAND" {
				detectingRedis = false
				session.SetMode(ModeRedis)
				session.Write(successMessage)
				continue
			}

			if session.mode == ModeRedis {
				// buffer
				var flush = false

				// @todo add and validate, if broken, reset
				if expectingTuples == 0 && len(lineTokens) == 0 {
					if strings.HasPrefix(line, "*") {
						n, _ := strconv.ParseUint(line[1:], 10, 64)
						if n < 1 {
							panic("should be >= 1")
						}
						expectingTuples = uint(n)
						continue
					} else {
						panic("first should be *x")
					}
				}
				// @todo later, we could use the length here to immediately read the length (e.g. $7 and read 7 next bytes that may contain []byte("COMMAND") or similar )
				if strings.HasPrefix(line, "$") {
					// ignore length markers
					continue
				}

				// append token
				lineTokens = append(lineTokens, line)

				// we have all tokens
				if uint(len(lineTokens)) == expectingTuples {
					flush = true
				}

				// wait for more tokens?
				if !flush {
					continue
				}

				// join as if it were a regular line
				line = strings.Join(lineTokens, " ")

				// reset
				lineTokens = make([]string, 0)
				expectingTuples = 0
			}

			// handle
			if err := session.Handle(InputLine(line)); err != nil {
				// pass error
				if err := session.WriteErrMessage(err); err != nil {
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

		// error?
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}
	}
}

func (instance *Instance) ServeTELNET(ctx tel.Context, w tel.Writer, r tel.Reader) {
	instance.Serve(w, r)
}

func New(opts *Opts) *Instance {
	return &Instance{
		opts: opts,
	}
}
