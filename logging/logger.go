package logging

import (
	"time"

	"github.com/hashicorp/go-hclog"
)

var statsPrefix = "service.emojify"

// Logger defines an interface for common logging operations
type Logger interface {
	Log() hclog.Logger

	ServiceStart(address, port, version string)

	CacheCheck() Finished
	CacheExists(string) Finished
	CacheGet(string) Finished
	CachePut(string) Finished

	CacheInvalidate() Finished
	CacheInvalidateItem(string, time.Duration, time.Duration, error)
}

// Finished defines a function to be returned by logging methods which contain timers
type Finished func(status int, err error)
