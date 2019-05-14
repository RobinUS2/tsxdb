package main

import (
	"github.com/RobinUS2/tsxdb/server"
	"log"
	"os"
	"os/signal"
)

var shutdown = make(chan os.Signal, 1)

func main() {
	opts := server.NewOpts()
	// @todo read config yaml
	opts.AuthToken = "123"

	// new instance
	instance := server.New(opts)

	// init
	if err := instance.Init(); err != nil {
		log.Fatalf("unable to initialise server %s", err)
	}

	// start
	if err := instance.Start(); err != nil {
		log.Fatalf("unable to start server %s", err)
	}
	log.Println("server started")

	// listen for shutdown
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
}
