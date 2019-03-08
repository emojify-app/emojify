package server

import (
	"fmt"
	"net"

	"github.com/emojify-app/cache/storage"
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/protos/emojify"
	"google.golang.org/grpc"
)

var lis net.Listener

var grpcServer *grpc.Server

// Start a new instance of the server
func Start(address string, port int, l logging.Logger, s storage.Store) error {
	grpcServer = grpc.NewServer()
	emojify.RegisterEmojifyServer(grpcServer, &Emojify{})

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
