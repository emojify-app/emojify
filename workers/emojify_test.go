package workers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/emojify"
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/queue"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/mock"
)

type testData struct {
	emo              *Emojify
	popChan          chan queue.PopResponse
	qi               queue.PopResponse
	mockQueue        *queue.MockQueue
	mockCache        *cache.ClientMock
	mockFetcher      *emojify.MockFetcher
	mockEmojify      *emojify.MockEmojify
	mockReader       *bytes.Reader
	mockFaces        []image.Rectangle
	mockImage        image.Image
	mockEmojifyImage image.Image
}

func setup(t *testing.T, timeout time.Duration) *testData {
	td := &testData{}
	td.popChan = make(chan queue.PopResponse)
	td.qi = queue.PopResponse{
		Item: &queue.Item{
			ID:  "abc123",
			URI: "https://something",
		},
		Error: nil,
	}

	td.mockQueue = &queue.MockQueue{}
	td.mockQueue.On("Pop").Return(td.popChan)

	td.mockReader = bytes.NewReader([]byte("abc"))
	td.mockFaces = []image.Rectangle{image.Rect(0, 0, 10, 10)}
	td.mockImage = image.NewUniform(color.Black)
	td.mockEmojifyImage = image.NewRGBA64(image.Rect(0, 0, 400, 400))

	td.mockCache = &cache.ClientMock{}
	td.mockCache.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(&wrappers.BoolValue{Value: false}, nil)
	td.mockCache.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(&wrappers.StringValue{Value: "abc"}, nil)

	td.mockFetcher = &emojify.MockFetcher{}
	td.mockFetcher.On("FetchImage", mock.Anything).Return(td.mockReader, nil)
	td.mockFetcher.On("ReaderToImage", td.mockReader).Return(td.mockImage, nil)

	td.mockEmojify = &emojify.MockEmojify{}
	td.mockEmojify.On("GetFaces", td.mockReader).Return(td.mockFaces, nil)
	td.mockEmojify.On("Emojimise", td.mockImage, td.mockFaces).Return(td.mockEmojifyImage, nil)

	logger := logging.New("localhost:9125", "debug")

	td.emo = &Emojify{
		queue:       td.mockQueue,
		cache:       td.mockCache,
		logger:      logger,
		fetcher:     td.mockFetcher,
		emojifier:   td.mockEmojify,
		errorDelay:  1 * time.Millisecond,
		normalDelay: 1 * time.Millisecond}
	go td.emo.Start() // start the app

	return td
}

func TestStartWithCacheItemDoesNotFetch(t *testing.T) {
	td := setup(t, 10*time.Millisecond)
	id := &wrappers.StringValue{Value: "abc123"}

	td.mockCache.ExpectedCalls = make([]*mock.Call, 0)
	td.mockCache.On("Exists", mock.Anything, id, mock.Anything).Return(&wrappers.BoolValue{Value: true}, nil)

	td.popChan <- td.qi
	time.Sleep(1000 * time.Millisecond)

	td.mockCache.AssertCalled(t, "Exists", mock.Anything, id, mock.Anything)
	td.mockFetcher.AssertNotCalled(t, "FetchImage", mock.Anything)
}

func TestStartWithCacheErrorDoesNotFetch(t *testing.T) {
	td := setup(t, 10*time.Millisecond)
	id := &wrappers.StringValue{Value: "abc123"}

	td.mockCache.ExpectedCalls = make([]*mock.Call, 0)
	td.mockCache.On("Exists", mock.Anything, id, mock.Anything).Return(nil, fmt.Errorf("abc"))

	td.popChan <- td.qi
	time.Sleep(1000 * time.Millisecond)

	td.mockCache.AssertCalled(t, "Exists", mock.Anything, id, mock.Anything)
	td.mockFetcher.AssertNotCalled(t, "FetchImage", mock.Anything)
}

func TestStartWithFetchErrorDoesNotFindFaces(t *testing.T) {
	td := setup(t, 10*time.Millisecond)

	td.mockFetcher.ExpectedCalls = make([]*mock.Call, 0)
	td.mockFetcher.On("FetchImage", mock.Anything).Return(nil, fmt.Errorf("abc"))

	td.popChan <- td.qi
	time.Sleep(1000 * time.Millisecond)

	td.mockEmojify.AssertNotCalled(t, "GetFaces", mock.Anything)
}

func TestStartWithInvalidImageDoesNotFindFaces(t *testing.T) {
	td := setup(t, 10*time.Millisecond)

	td.mockFetcher.ExpectedCalls = make([]*mock.Call, 0)
	td.mockFetcher.On("FetchImage", mock.Anything).Return(td.mockReader, nil)
	td.mockFetcher.On("ReaderToImage", mock.Anything).Return(nil, fmt.Errorf("abc"))

	td.popChan <- td.qi
	time.Sleep(1000 * time.Millisecond)

	td.mockEmojify.AssertNotCalled(t, "GetFaces", mock.Anything)
}

func TestStartWithInvalidEmojimiseDoesNotSetCache(t *testing.T) {
	td := setup(t, 10*time.Millisecond)

	td.mockEmojify.ExpectedCalls = make([]*mock.Call, 0)
	td.mockEmojify.On("GetFaces", td.mockReader).Return(td.mockFaces, nil)
	td.mockEmojify.On("Emojimise", td.mockImage, td.mockFaces).Return(nil, fmt.Errorf("boom"))

	td.popChan <- td.qi
	time.Sleep(1000 * time.Millisecond)

	td.mockEmojify.AssertCalled(t, "Emojimise", td.mockImage, td.mockFaces)
	td.mockCache.AssertNotCalled(t, "Put", mock.Anything)
}

func TestStartProcessesItem(t *testing.T) {
	td := setup(t, 10*time.Millisecond)
	id := &wrappers.StringValue{Value: "abc123"}

	td.popChan <- td.qi
	time.Sleep(1000 * time.Millisecond)

	td.mockCache.AssertCalled(t, "Exists", mock.Anything, id, mock.Anything)
	td.mockFetcher.AssertCalled(t, "FetchImage", td.qi.Item.URI)
	td.mockEmojify.AssertCalled(t, "GetFaces", td.mockReader)
	td.mockEmojify.AssertCalled(t, "Emojimise", td.mockImage, td.mockFaces)
	td.mockCache.AssertCalled(t, "Put", mock.Anything, mock.Anything, mock.Anything)
}
