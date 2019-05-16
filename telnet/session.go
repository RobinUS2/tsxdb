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
	_ = session.Write(line)

	return nil
}

func (session *Session) SetWriter(writer tel.Writer) {
	session.writer = writer
}

func (session *Session) Write(s string) error {
	if !strings.HasSuffix(s, "\n") {
		s = s + "\n"
	}
	b := []byte(s)
	if nWritten, err := oi.LongWrite(session.writer, b); err != nil || int64(len(b)) != nWritten {
		return err
	}
	return nil
}

func NewSession(instance *Instance) *Session {
	return &Session{
		instance: instance,
	}
}

type InputLine string
