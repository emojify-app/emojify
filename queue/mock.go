package queue

import "github.com/stretchr/testify/mock"

// MockQueue is a mock implementation of the queue interface for testing
type MockQueue struct {
	mock.Mock
}

// Push an item onto the queue
func (q *MockQueue) Push(i *Item) (int, int, error) {

	args := q.Called(i)

	return args.Get(0).(int), args.Get(1).(int), args.Error(2)
}

// Pop the last item off the queue
func (q *MockQueue) Pop() chan PopResponse {
	args := q.Called()

	return args.Get(0).(chan PopResponse)
}

// Position allows you to query the position of an item in the queue
func (q *MockQueue) Position(key string) (position, length int, err error) {
	args := q.Called(key)

	return args.Get(0).(int), args.Get(1).(int), args.Error(2)
}

// Ping is a mock implementation of the the Ping function
func (q *MockQueue) Ping() error {
	args := q.Called()
	return args.Error(0)
}
