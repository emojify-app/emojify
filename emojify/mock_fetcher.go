package emojify

import (
	"image"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) FetchImage(uri string) (io.ReadSeeker, error) {
	args := m.Called(uri)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(io.ReadSeeker), args.Error(1)
}

func (m *MockFetcher) ReaderToImage(r io.ReadSeeker) (image.Image, error) {
	args := m.Called(r)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(image.Image), args.Error(1)
}
