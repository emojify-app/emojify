package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/emojify"
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/queue"
	"github.com/emojify-app/emojify/server"
	"github.com/emojify-app/emojify/workers"
	"github.com/emojify-app/face-detection/client"
	"github.com/nicholasjackson/env"
	"google.golang.org/grpc"
)

var version = "dev"

var envBindAddress = env.String("BIND_ADDRESS", false, "localhost", "Bind address for gRPC server, e.g. 127.0.0.1")
var envBindPort = env.Integer("BIND_PORT", false, 9090, "Bind port for gRPC server e.g. 9090")

var envHealthBindAddress = env.String("HEALTH_BIND_ADDRESS", false, "localhost", "Bind address for health endpoint, e.g. 127.0.0.1")
var envHealthBindPort = env.Integer("HEALTH_BIND_PORT", false, 9091, "Bind port for health endpoint e.g. 9091")

var redisAddress = env.String("REDIS_ADDRESS", false, "localhost:6379", "Address for redis server")
var redisPassword = env.String("REDIS_PASSWORD", false, "", "Password for redis server")
var redisDB = env.Integer("REDIS_DB", false, 0, "Database for redis server")

var cacheAddress = env.String("CACHE_ADDRESS", false, "localhost:8000", "Address for cache server")

var faceboxAddress = env.String("FACEBOX_ADDRESS", false, "localhost:8001", "Address for facebox server")

var statsDAddress = env.String("STATSD_ADDRESS", false, "localhost:8125", "Address for statsd server")
var logLevel = env.String("LOG_LEVEL", false, "info", "Level for log output [info,debug,trace,error]")

var help = flag.Bool("help", false, "--help to show help")

func main() {
	flag.Parse()
	flag.Parsed()

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

	// setup dependencies
	l := logging.New(*statsDAddress, *logLevel)

	q, err := queue.New(*redisAddress, *redisPassword, *redisDB)
	if err != nil {
		l.Log().Error("Unable to create queue", err)
		os.Exit(1)
	}

	conn, err := grpc.Dial(*cacheAddress, grpc.WithInsecure())
	if err != nil {
		l.Log().Error("Unable to create gRPC client", err)
		os.Exit(1)
	}
	cc := cache.NewCacheClient(conn)

	f := emojify.NewFetcher()
	fd := client.NewClient(*faceboxAddress)
	e, err := emojify.NewEmojify("./images/", fd)
	if err != nil {
		l.Log().Error("Unable to load emojies", err)
		os.Exit(1)
	}

	w := workers.New(q, cc, l, f, e, 30*time.Second, 100*time.Millisecond)
	go w.Start() // start the worker and process queue items

	http.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(fmt.Sprintf("%s:%d", *envHealthBindAddress, *envHealthBindPort), nil)

	l.Log().Info("Binding gRPC to", "address", *envBindAddress, "port", *envBindPort)
	l.Log().Info("Starting gRPC server")

	err = server.Start(*envBindAddress, *envBindPort, l, cc, q)
	if err != nil {
		l.Log().Error("Unable to start server", "error", err)
		os.Exit(1)
	}
}
