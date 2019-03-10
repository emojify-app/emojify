package logging

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/emojify-app/emojify/queue"
	"github.com/hashicorp/go-hclog"
)

var statsPrefix = "service.emojify"

// Logger defines an interface for common logging operations
type Logger interface {
	Log() hclog.Logger

	ServiceStart(address, port, version string)

	// gRPC Endpoint logging
	Create(string) Finished
	Query(string) Finished

	// Cache Operations
	CacheExists(string) Finished
	CachePut(string) Finished

	// Queue Operations
	QueueGet(string) Finished
	QueuePut(string) Finished

	// Emojify Worker
	WorkerProcessQueueItem(*queue.Item) Finished
	WorkerFetchImage(uri string) Finished
	WorkerInvalidImage(uri string, err error)
	WorkerFindFaces(uri string) Finished
	WorkerEmojify(uri string) Finished
	WorkerImageEncodeError(uri string, err error)
}

// Finished defines a function to be returned by logging methods which contain timers
type Finished func(status int, err error)

// Impl is a concrete implementation of the Logger interface
type Impl struct {
	l hclog.Logger
	s *statsd.Client
}

// New creates a new logger implementation
func New(statsdAddress string) Logger {
	l := hclog.Default()
	s, err := statsd.New(statsdAddress)
	if err != nil {
		panic(err)
	}

	return &Impl{l, s}
}

// Log returns the raw logger for arbitary messages
func (i *Impl) Log() hclog.Logger {
	return i.l
}

// ServiceStart logs information when the service starts
func (i *Impl) ServiceStart(address, port, version string) {
	i.s.Incr(statsPrefix+"service_start", nil, 1)
	i.l.Info("Emojify service started", "address", address, "port", port, "version", version)
}

// Create logs timing information related to the gRPC Create method
func (i *Impl) Create(uri string) Finished {
	st := time.Now()
	i.l.Debug("Create called", "uri", uri)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+".create", time.Now().Sub(st), getStatusTags(status), 1)

		if err != nil {
			i.l.Error("Create error", "uri", uri, "status", status, "error", err)
			return
		}

		i.l.Debug("Create finished", "uri", uri, "status", status)
	}
}

// Query logs timing information related to the gRPC Query method
func (i *Impl) Query(key string) Finished {
	st := time.Now()
	i.l.Debug("Query called", "key", key)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+".query", time.Now().Sub(st), getStatusTags(status), 1)

		if err != nil {
			i.l.Error("Query error", "key", key, "status", status, "error", err)
			return
		}

		i.l.Debug("Query finished", "key", key, "status", status)
	}

}

// CacheExists logs timing information related to Cache service exists method calls
func (i *Impl) CacheExists(key string) Finished {
	st := time.Now()
	i.l.Debug("Check cache called", "key", key)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+".cache.exists", time.Now().Sub(st), getStatusTags(status), 1)

		if err != nil {
			i.l.Error("Cache check error", "key", key, "status", status, "error", err)
			return
		}

		i.l.Debug("Cache check finished", "key", key, "status", status)
	}
}

// CachePut logs information when an image is pushed to the cache
func (i *Impl) CachePut(key string) Finished {
	st := time.Now()
	i.l.Debug("Cache image", "key", key)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+"cache.put", time.Now().Sub(st), getStatusTags(status), 1)
		i.l.Debug("Cache image finished", "key", key, "status", status)

		if err != nil {
			i.l.Error("Unable to save image to cache", "key", key, "error", err)
		}
	}
}

// QueueGet logs timing information related to querying the status of an item on the queue
func (i *Impl) QueueGet(key string) Finished {
	st := time.Now()
	i.l.Debug("Queue query called", "key", key)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+".queue.query", time.Now().Sub(st), getStatusTags(status), 1)

		if err != nil {
			i.l.Error("Queue query error", "key", key, "status", status, "error", err)
			return
		}

		i.l.Debug("Queue query finished", "key", key, "status", status)
	}
}

// QueuePut logs timing information when an item is stored on the queue
func (i *Impl) QueuePut(key string) Finished {
	st := time.Now()
	i.l.Debug("Queue put called", "key", key)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+".queue.put", time.Now().Sub(st), getStatusTags(status), 1)

		if err != nil {
			i.l.Error("Queue put error", "key", key, "status", status, "error", err)
			return
		}

		i.l.Debug("Queue put finished", "key", key, "status", status)
	}
}

// WorkerProcessQueueItem logs information about the processing of a queue item
func (i *Impl) WorkerProcessQueueItem(item *queue.Item) Finished {
	st := time.Now()
	i.l.Debug("Processing queue item", "item", item)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+"worker.process", time.Now().Sub(st), getStatusTags(status), 1)
		i.l.Debug("Processing queue item finished", "item", item, "status", status)

		if err != nil {
			i.l.Error("Error processing queue item", "status", status, "item", item, "error", err)
		}
	}
}

// WorkerFetchImage logs information about a remote fetch for the image
func (i *Impl) WorkerFetchImage(uri string) Finished {
	st := time.Now()
	i.l.Debug("Fetching file", "uri", uri)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+"worker.fetch_file", time.Now().Sub(st), getStatusTags(status), 1)
		i.l.Debug("Fetching file finished", "uri", uri, "status", status)

		if err != nil {
			i.l.Error("Error fetching file", "status", status, "uri", uri, "error", err)
		}
	}
}

// WorkerInvalidImage logs information when an invalid image is returned from the fetch
func (i *Impl) WorkerInvalidImage(uri string, err error) {
	i.l.Error("Invalid image format", "uri", uri, "error", err)
	i.s.Incr(statsPrefix+"worker.invalid_image", nil, 1)
}

// WorkerFindFaces logs information related to the face lookup call
func (i *Impl) WorkerFindFaces(uri string) Finished {
	st := time.Now()
	i.l.Debug("Find faces in image", "uri", uri)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+"worker.find_faces", time.Now().Sub(st), getStatusTags(status), 1)
		i.l.Debug("Find faces finished", "uri", uri, "status", status)

		if err != nil {
			i.l.Error("Unable to find faces", "handler", "emojify", "uri", uri, "error", err)
		}
	}
}

// WorkerEmojify logs information when emojifying the image
func (i *Impl) WorkerEmojify(uri string) Finished {
	st := time.Now()
	i.l.Debug("Emojify image", "uri", uri)

	return func(status int, err error) {
		i.s.Timing(statsPrefix+"worker.find_faces", time.Now().Sub(st), getStatusTags(status), 1)
		i.l.Debug("Find faces finished", "uri", uri, "status", status)

		if err != nil {
			i.l.Error("Unable to emojify", "uri", uri, "error", err)
		}
	}
}

// WorkerImageEncodeError logs information when an image encode error occurs
func (i *Impl) WorkerImageEncodeError(uri string, err error) {
	i.l.Error("Unable to encode file as jpg", "uri", uri, "error", err)
	i.s.Incr(statsPrefix+"worker.image_encode_error", nil, 1)
}

func getStatusTags(status int) []string {
	return []string{
		fmt.Sprintf("status:%d", status),
	}
}
