package server

import (
	"fmt"
	"net"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/emojify-app/emojify/queue"
	"google.golang.org/grpc"
)

var lis net.Listener

var grpcServer *grpc.Server

// Start a new instance of the server
func Start(address string, port int, l logging.Logger, c cache.CacheClient, q queue.Queue) error {
	grpcServer = grpc.NewServer()
	emojify.RegisterEmojifyServer(grpcServer, &Emojify{q, c, l})

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return err
	}

	return grpcServer.Serve(lis)
}

// Stop the server
func Stop() error {
	grpcServer.Stop()
	//	lis.Close()
	return nil
}
