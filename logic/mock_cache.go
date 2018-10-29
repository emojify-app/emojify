package logic

import (
	"github.com/stretchr/testify/mock"
)

// MockCache is a mock implementation of the Cache interface for testing
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Exists(key string) (bool, error) {
	args := m.Called(key)

	return args.Bool(0), args.Error(1)
}

// Get calls the mock get function
func (m *MockCache) Get(key string) ([]byte, error) {
	args := m.Called(key)

	return args.Get(0).([]byte), args.Error(1)
}

// Put calls the mock put function
func (m *MockCache) Put(key string, data []byte) error {
	args := m.Called(key)
	return args.Error(0)
}
