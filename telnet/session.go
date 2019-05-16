package telnet

import (
	"github.com/reiver/go-oi"
	tel "github.com/reiver/go-telnet"
	"log"
)

type Session struct {
	writer tel.Writer
}

func (session *Session) Handle(line InputLine) error {
	log.Println(line)

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

func NewSession() *Session {
	return &Session{}
}

type InputLine string
