package logging

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/statsd"
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
	CacheExists(string) Finished

	// Queue Operations
	QueueGet(string) Finished
	QueuePut(string) Finished
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

func getStatusTags(status int) []string {
	return []string{
		fmt.Sprintf("status:%d", status),
	}
}
