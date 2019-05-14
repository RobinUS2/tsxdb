package main

import (
	"github.com/RobinUS2/tsxdb/server"
	"log"
)

func main() {
	opts := server.NewOpts()
	opts.AuthToken = "123"
	instance := server.New(opts)
	if err := instance.Init(); err != nil {
		log.Fatalf("unable to initialise server %s", err)
	}
	if err := instance.Start(); err != nil {
		log.Fatalf("unable to start server %s", err)
	}
}
