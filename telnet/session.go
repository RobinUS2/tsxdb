package telnet

import (
	"errors"
	"github.com/reiver/go-oi"
	tel "github.com/reiver/go-telnet"
	"log"
	"strings"
)

type Session struct {
	instance      *Instance
	writer        tel.Writer
	authenticated bool
}

func (session *Session) Handle(typedLine InputLine) error {
	line := string(typedLine)
	log.Printf("telnet rcv %s", line)

	// auth
	if strings.HasPrefix(line, "auth ") {
		token := strings.SplitN(line, "auth ", 2)[1]
		if token != session.instance.opts.AuthToken {
			return errors.New("invalid auth token")
		}
		// good
		session.authenticated = true
	}

	// authenticated?
	if !session.authenticated {
		return errors.New("not authenticated")
	}

	// echo back
	b := []byte(line + "\n")
	n := len(b)
	if nWritten, err := oi.LongWrite(session.writer, b); err != nil || nWritten != int64(n) {
		panic("failed to write")
	}

	return nil
}

func (session *Session) SetWriter(writer tel.Writer) {
	session.writer = writer
}

func NewSession(instance *Instance) *Session {
	return &Session{
		instance: instance,
	}
}

type InputLine string
