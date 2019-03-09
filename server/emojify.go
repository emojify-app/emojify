package server

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/logging"
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
	logger      logging.Logger
}

// New creates a new Emojify implementation
func New(q queue.Queue, cc cache.CacheClient, l logging.Logger) *Emojify {
	return &Emojify{q, cc, l}
}

// Check is a gRPC health check
func (e *Emojify) Check(context.Context, *emojify.HealthCheckRequest) (*emojify.HealthCheckResponse, error) {
	return nil, nil
}

// Create an Emojify request to process an image
func (e *Emojify) Create(ctx context.Context, uri *wrappers.StringValue) (*emojify.QueryItem, error) {
	done := e.logger.Create(uri.GetValue())

	id := base64.URLEncoding.EncodeToString([]byte(uri.GetValue()))

	// check the current queue and cache before adding
	ei, err := e.checkQueueAndCache(id)
	if ei != nil || err != nil {
		done(http.StatusInternalServerError, err)
		return ei, err
	}

	// create a new query item
	ei = &emojify.QueryItem{
		Id:     id,
		Status: &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED},
	}

	// create a new queueItem
	qi := &queue.Item{
		ID:    id,
		Added: time.Now(),
	}

	queueDone := e.logger.QueuePut(id)
	err = e.workerQueue.Push(qi)
	if err != nil {
		queueDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, err)
		return nil, grpc.Errorf(codes.Internal, "error addding to queue: %s", err)
	}

	queueDone(http.StatusOK, nil)
	done(http.StatusOK, nil)
	return ei, nil
}

// Query the status of an Emojify request
func (e *Emojify) Query(ctx context.Context, id *wrappers.StringValue) (*emojify.QueryItem, error) {
	done := e.logger.Query(id.GetValue())

	ei, err := e.checkQueueAndCache(id.GetValue())
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			done(http.StatusNotFound, err)
		} else {
			done(http.StatusInternalServerError, err)
		}

		return ei, err
	}

	done(http.StatusOK, nil)
	return ei, nil
}

func (e *Emojify) checkQueueAndCache(id string) (*emojify.QueryItem, error) {
	ei := &emojify.QueryItem{Id: id}

	cDone := e.logger.CacheExists(id)
	// check the item is not all ready cached
	ok, err := e.cache.Exists(context.Background(), &wrappers.StringValue{Value: id})
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			cDone(http.StatusNotFound, err)
			return nil, grpc.Errorf(codes.NotFound, "cache item not found: %s", err)
		}

		cDone(http.StatusInternalServerError, err)
		return nil, grpc.Errorf(codes.Internal, "cache error: %s", err)
	}

	if ok.GetValue() {
		cDone(http.StatusOK, nil)

		ei.Status = &emojify.QueryStatus{Status: emojify.QueryStatus_FINISHED}
		return ei, nil
	}

	qiDone := e.logger.QueueGet(id)
	// check the item is not already on the queue
	pos, l, err := e.workerQueue.Position(id)
	if err != nil {
		qiDone(http.StatusInternalServerError, err)
		return nil, grpc.Errorf(codes.Internal, "error getting position: %s", err)
	}

	if pos > 0 {
		ei.QueuePosition = int32(pos)
		ei.QueueLength = int32(l)
		ei.Status = &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED}

		qiDone(http.StatusOK, nil)
		return ei, nil
	}

	qiDone(http.StatusNotFound, nil)
	return nil, nil
}
