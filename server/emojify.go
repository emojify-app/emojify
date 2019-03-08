package server

import (
	"context"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/emojify-app/emojify/queue"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Emojify implements the gRPC server interface methods
type Emojify struct {
	workerQueue queue.Queue
	cache       cache.CacheClient
}

// New creates a new Emojify implementation
func New(q queue.Queue, cc cache.CacheClient) *Emojify {
	return &Emojify{q, cc}
}

// Check is a gRPC health check
func (e *Emojify) Check(context.Context, *emojify.HealthCheckRequest) (*emojify.HealthCheckResponse, error) {
	return nil, nil
}

// Create an Emojify request to process an image
func (e *Emojify) Create(ctx context.Context, id *wrappers.StringValue) (*emojify.QueryItem, error) {
	qi := &queue.Item{ID: id.GetValue()}
	ei := &emojify.QueryItem{Id: qi.ID}

	// check the item is not all ready cached
	ok, err := e.cache.Exists(context.Background(), id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "cache error: %s", err)
	}

	if ok.GetValue() {
		ei.Status = &emojify.QueryStatus{Status: emojify.QueryStatus_FINISHED}
		return ei, nil
	}

	// check the item is not already on the queue
	pos, l, _ := e.workerQueue.Position(id.GetValue())
	if pos > 0 {
		ei.QueuePosition = int32(pos)
		ei.QueueLength = int32(l)
		ei.Status = &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED}
		return ei, nil
	}

	e.workerQueue.Push(qi)
	ei.Status = &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED}

	return ei, nil
}

// Query the status of an Emojify request
func (e *Emojify) Query(ctx context.Context, id *wrappers.StringValue) (*emojify.QueryItem, error) {
	return nil, nil
}
