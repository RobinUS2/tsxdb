package telnet

import (
	"github.com/reiver/go-oi"
	tel "github.com/reiver/go-telnet"
	"log"
	"strings"
)

type Session struct {
	instance *Instance
	writer   tel.Writer
}

func (session *Session) Handle(typedLine InputLine) error {
	line := string(typedLine)
	log.Println(line)

	if strings.HasPrefix(line, "auth ") {
		token := strings.SplitN(line, "auth ", 2)
		log.Println(token[1])
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
