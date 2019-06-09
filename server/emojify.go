package server

import (
	"context"
	"encoding/base64"
	"log"
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
	resp := emojify.HealthCheckResponse{}

	// is redis connected
	err := e.workerQueue.Ping()
	if err != nil {
		resp.Status = emojify.HealthCheckResponse_NOT_SERVING
		return &resp, err
	}

	resp.Status = emojify.HealthCheckResponse_SERVING

	return &resp, nil
}

// Create an Emojify request to process an image
func (e *Emojify) Create(ctx context.Context, uri *wrappers.StringValue) (*emojify.QueryItem, error) {
	done := e.logger.Create(uri.GetValue())

	id := base64.URLEncoding.EncodeToString([]byte(uri.GetValue()))

	// check the current queue and cache before adding
	ei, err := e.checkQueueAndCache(id)
	if ei == nil || err != nil {
		e.logger.Log().Debug("create finished with 500", "error", err)
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
		URI:   uri.GetValue(),
	}

	e.logger.Log().Debug("create put")
	queueDone := e.logger.QueuePut(id)
	pos, length, err := e.workerQueue.Push(qi)
	if err != nil {
		queueDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, err)
		e.logger.Log().Debug("create finished with 500")
		return nil, grpc.Errorf(codes.Internal, "error addding to queue: %s", err)
	}
	queueDone(http.StatusOK, nil)

	e.logger.WorkerQueueStatus(length)

	ei.QueuePosition = int32(pos)
	ei.QueueLength = int32(length)

	e.logger.Log().Debug("create finished with 200")
	done(http.StatusOK, nil)
	return ei, nil
}

// Query the status of an Emojify request
func (e *Emojify) Query(ctx context.Context, id *wrappers.StringValue) (*emojify.QueryItem, error) {
	done := e.logger.Query(id.GetValue())

	ei, err := e.checkQueueAndCache(id.GetValue())
	if err != nil {
		log.Println(err)
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
	// check the item is not all ready cached returns ok if found in cache
	ok, err := e.cache.Exists(context.Background(), &wrappers.StringValue{Value: id})
	if err != nil {
		e.logger.Log().Error("error cache", "err", err)
		if grpc.Code(err) == codes.NotFound {
			cDone(http.StatusNotFound, err)
			return ei, grpc.Errorf(codes.NotFound, "cache item not found: %s", err)
		}

		cDone(http.StatusInternalServerError, err)
		return ei, grpc.Errorf(codes.Internal, "cache error: %s", err)
	}

	// found item in the cache return finished
	if ok.GetValue() == true {
		e.logger.Log().Error("error found", "ok", ok)
		cDone(http.StatusOK, nil)

		ei.Status = &emojify.QueryStatus{Status: emojify.QueryStatus_FINISHED}
		return ei, nil
	}

	// check the item is not already on the queue do not return an error
	// as the queue might not exist
	qiDone := e.logger.QueueGet(id)
	pos, l, err := e.workerQueue.Position(id)
	if err != nil {
		qiDone(http.StatusInternalServerError, err)
		return ei, nil
	}

	if pos > 0 {
		ei.QueuePosition = int32(pos)
		ei.QueueLength = int32(l)
		ei.Status = &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED}

		qiDone(http.StatusOK, nil)
		return ei, nil
	}

	qiDone(http.StatusNotFound, nil)
	return ei, nil
}
