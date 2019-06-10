package emojify

import (
	"image"
	"io"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockEmojify is a mock implementation of the Emojify interface
type MockEmojify struct {
	mock.Mock
}

// Emojimise is a mock implementation of the interface function
func (m *MockEmojify) Emojimise(src image.Image, faces []image.Rectangle) (image.Image, error) {
	args := m.Called(src, faces)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(image.Image), args.Error(1)
}

// GetFaces is a mock implementation of the interface function
func (m *MockEmojify) GetFaces(r io.ReadSeeker) ([]image.Rectangle, error) {
	args := m.Called(r)

	// wait for the client to block
	time.Sleep(10 * time.Millisecond)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]image.Rectangle), args.Error(1)
}

// Health is a mock implementation of the interface function
func (m *MockEmojify) Health() (int, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return 0, args.Error(1)
	}

	return args.Get(0).(int), args.Error(1)
}
