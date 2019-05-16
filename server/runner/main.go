package main

import (
	"flag"
	"github.com/RobinUS2/tsxdb/server"
	"github.com/RobinUS2/tsxdb/tools"
	"log"
	"os"
	"os/signal"
)

var shutdown = make(chan os.Signal, 1)

var configPathsStr string

func init() {
	flag.StringVar(&configPathsStr, "config", "config.yaml,/etc/tsxdb/config.yaml", "Configuration file (path(s))")
	flag.Parse()
}

func main() {
	// config
	opts := server.NewOpts()

	// read config
	if err := tools.ReadYamlFileInPath(configPathsStr, opts); err != nil {
		log.Fatalf("failed to read config %s", err)
	}

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
	log.Printf("server started %s:%d", opts.ListenHost, opts.ListenPort)

	// listen for shutdown
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
}
