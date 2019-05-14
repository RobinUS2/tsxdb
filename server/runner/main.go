package main

import (
	"flag"
	"github.com/RobinUS2/tsxdb/server"
	"github.com/RobinUS2/tsxdb/tools"
	"log"
	"os"
	"os/signal"
	"strings"
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

	// read first config
	configRead := false
	configPaths := strings.Split(configPathsStr, ",")
	for _, configPath := range configPaths {
		if !tools.FileExists(configPath) {
			continue
		}
		if err := opts.ReadYamlFile(configPath); err != nil {
			log.Fatalf("unable to read config in %s: %s", configPath, err)
		}
		configRead = true
	}
	if !configRead {
		log.Fatalf("no config files found in %v", configPaths)
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
