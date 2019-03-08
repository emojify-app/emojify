package queue

import "time"

// Item defines an item which exists on the queue
type Item struct {
	// ID of the queue item
	ID string
	// URI of the item to process
	URI string
	// Added to the queue at time
	Added time.Time
	// Complete at time
	Complete time.Time
	// Retry count
	Retry int
	// Error, only set when processing error occurs
	Error error
}

// Queue defines the interface methods for a FIFO queue
type Queue interface {
	// Push an item onto the queue
	Push(*Item) error
	// Pop the last item off the queue
	Pop() (*Item, error)
	// Position allows you to query the position of an item in the queue
	Position(key string) (position, length int, err error)
}
