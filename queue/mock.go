package queue

import "github.com/stretchr/testify/mock"

type MockQueue struct {
	mock.Mock
}

// Push an item onto the queue
func (q *MockQueue) Push(Item) error {
	return nil
}

// Pop the last item off the queue
func (q *MockQueue) Pop() (*Item, error) {
	return nil, nil
}

// Position allows you to query the position of an item in the queue
func (q *MockQueue) Position(key string) (position, length int, err error) {
	return 0, 0, nil
}
