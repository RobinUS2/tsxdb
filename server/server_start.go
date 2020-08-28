package server

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/telnet"
	"log"
)

// start server listening, this should only be called once per instance
func (instance *Instance) Start() (err error) {
	// catch runtime errors
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("server runtime error %s", r)
		}
	}()

	// start server
	if err := instance.StartListening(); err != nil {
		return err
	}

	// telnet server
	if instance.Opts().TelnetPort > 0 {
		telOpts := telnet.NewOpts()
		telOpts.Port = instance.Opts().TelnetPort
		telOpts.Host = instance.Opts().TelnetHost
		telOpts.AuthToken = instance.Opts().AuthToken
		telOpts.ServerHost = instance.Opts().ListenHost
		telOpts.ServerPort = instance.Opts().ListenPort
		instance.telnetServer = telnet.New(telOpts)
		go func() {
			err := instance.telnetServer.Listen()
			if err != nil {
				log.Printf("telnet failed to listen %s", err)
			}
		}()
	}

	return nil
}
