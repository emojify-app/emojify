package logic

import (
	"image"
	"io"

	"github.com/machinebox/sdk-go/facebox"
	"github.com/stretchr/testify/mock"
)

type MockEmojify struct {
	mock.Mock
}

func (m *MockEmojify) Emojimise(src image.Image, faces []facebox.Face) (image.Image, error) {
	args := m.Called(src, faces)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(image.Image), args.Error(1)
}

func (m *MockEmojify) GetFaces(r io.ReadSeeker) ([]facebox.Face, error) {
	args := m.Called(r)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]facebox.Face), args.Error(1)
}
