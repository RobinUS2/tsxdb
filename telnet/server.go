package telnet

import (
	"fmt"
	"github.com/reiver/go-telnet"
	"log"
)

type Instance struct {
	port int
}

func (instance *Instance) Listen() error {
	var handler = telnet.EchoHandler
	listenStr := fmt.Sprintf(":%d", instance.port)
	log.Printf("telnet listening at %s", listenStr)
	err := telnet.ListenAndServe(listenStr, handler)
	if nil != err {
		return err
	}
	return nil
}

func New(port int) *Instance {
	return &Instance{
		port: port,
	}
}
