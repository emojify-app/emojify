package workers

import (
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/queue"
)

type Emojify struct {
	queue  queue.Queue
	logger logging.Logger
}
