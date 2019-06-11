package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
)

var emojifyClient emojify.EmojifyClient

func main() {
	log.Println("Connecting to emojify", "address", "localhost:9090")
	emojifyConn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil {
		log.Println("Unable to create gRPC client", err)
		os.Exit(1)
	}
	emojifyClient = emojify.NewEmojifyClient(emojifyConn)

	resp, err := emojifyClient.Check(context.Background(), &emojify.HealthCheckRequest{})
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println(resp.String())
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go doWork("https://emojify.today/pictures/6.jpg", wg)
	go doWork("https://emojify.today/pictures/7.jpg", wg)

	wg.Wait()
}

func doWork(url string, wg *sync.WaitGroup) {
	// post an image to the server
	postresp, err := emojifyClient.Create(context.Background(), &wrappers.StringValue{Value: url})
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Create finished", postresp.String())

	if postresp.GetStatus().GetStatus() != emojify.QueryStatus_FINISHED {
		for {
			getresp, err := emojifyClient.Query(context.Background(), &wrappers.StringValue{Value: postresp.GetId()})
			if err != nil {
				log.Println("Get error", err)
			} else {
				log.Println("Get finished", getresp)
				if getresp.GetStatus().GetStatus() == emojify.QueryStatus_FINISHED {
					break
				}
			}

			time.Sleep(5 * time.Second)
		}
	}

	// fetch the image
	log.Println("Complete")

	wg.Done()
}
