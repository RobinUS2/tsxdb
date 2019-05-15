package telnet

import (
	"fmt"
	tel "github.com/reiver/go-telnet" // weird things happen if package with same name is imported as the package/module it's in unless aliased
	"log"
)

type Instance struct {
	port int
}

func (instance *Instance) Listen() error {
	listenStr := fmt.Sprintf(":%d", instance.port)
	log.Printf("telnet listening at %s", listenStr)
	err := tel.ListenAndServe(listenStr, instance)
	if nil != err {
		return err
	}
	return nil
}

func (instance *Instance) ServeTELNET(ctx tel.Context, w tel.Writer, r tel.Reader) {
	e := tel.EchoHandler
	// @todo implement
	e.ServeTELNET(ctx, w, r)
}

func New(port int) *Instance {
	return &Instance{
		port: port,
	}
}
