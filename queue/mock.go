package queue

import "github.com/stretchr/testify/mock"

// MockQueue is a mock implementation of the queue interface for testing
type MockQueue struct {
	mock.Mock
}

// Push an item onto the queue
func (q *MockQueue) Push(i *Item) error {

	args := q.Called(i)

	return args.Error(0)
}

// Pop the last item off the queue
func (q *MockQueue) Pop() (*Item, error) {
	return nil, nil
}

// Position allows you to query the position of an item in the queue
func (q *MockQueue) Position(key string) (position, length int, err error) {
	args := q.Called(key)

	return args.Get(0).(int), args.Get(1).(int), args.Error(2)
}
