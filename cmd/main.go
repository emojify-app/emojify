package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/emojify-app/emojify"
	"github.com/nicholasjackson/env"
)

var version = "v0.0.0"

var bindAddress = env.String("BIND_ADDR", false, "localhost:9090", "Bind address for the server, i.e. localhost:9090")
var cacheService = env.String("CACHE_SERVICE", true, "", "Name of the cache server as registered with consul, i.e. emojify-cache")
var consulAddress = env.String("CONSUL_ADDR", false, "http://localhost:8500", "Address of the local consul agent")
var serviceName = env.String("SERVICE_NAME", false, "emojify-service", "Name of this service to register in Consul")
var machineBoxAddress = env.String("MACHINEBOX_ADDR", true, "", "URI for the MachineBox server, i.e. http://localhost:9090")

var help = flag.Bool("help", false, "--help to show help")

func main() {
	//logger := hclog.Default()
	flag.Parse()

	// if the help flag is passed show configuration options
	if *help == true {
		fmt.Println("Emojify service version:", version)
		fmt.Println("Configuration values are set using environment variables, for info please see the following list")
		fmt.Println("")
		fmt.Println(env.Help())
	}

	err := env.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	server := emojify.NewServer(*bindAddress, *cacheService, *consulAddress, *machineBoxAddress, *serviceName, "./images")

	server.Start()
}
