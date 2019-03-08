package server

import (
	"context"

	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// Emojify implements the gRPC server interface methods
type Emojify struct{}

// Check is a gRPC health check
func (e *Emojify) Check(context.Context, *emojify.HealthCheckRequest) (*emojify.HealthCheckResponse, error) {
	return nil, nil
}

// Create an Emojify request to process an image
func (e *Emojify) Create(context.Context, *wrappers.StringValue) (*emojify.EmojifyItem, error) {
	return nil, nil
}

// Query the status of an Emojify request
func (e *Emojify) Query(context.Context, *emojify.EmojifyItem) (*emojify.QueryItem, error) {
	return nil, nil
}
