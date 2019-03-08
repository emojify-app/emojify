package server

import (
	"context"
	"testing"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/emojify-app/emojify/queue"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var mockQueue *queue.MockQueue
var mockCache *cache.ClientMock

func setup(t *testing.T, pos, ql int) *Emojify {
	mockQueue = &queue.MockQueue{}
	mockQueue.On("Push", mock.Anything).Return(nil)
	mockQueue.On("Position", mock.Anything).Return(pos, ql, nil)

	mockCache = &cache.ClientMock{}
	mockCache.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(&wrappers.BoolValue{Value: false}, nil)

	return New(mockQueue, mockCache)
}

func TestCreateAddsItemToTheQueueIfNotPresent(t *testing.T) {
	e := setup(t, 0, 0)
	id := &wrappers.StringValue{Value: "abc"}

	i, err := e.Create(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	mockQueue.AssertCalled(t, "Position", mock.Anything)
	mockQueue.AssertCalled(t, "Push", mock.Anything)
	assert.Equal(t, id.GetValue(), i.Id)
	assert.Equal(t, &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED}, i.GetStatus())
}

func TestCreateReturnsErrorWhenCacheError(t *testing.T) {
	e := setup(t, 0, 0)
	id := &wrappers.StringValue{Value: "abc"}
	mockCache.ExpectedCalls = make([]*mock.Call, 0)
	mockCache.On("Exists", mock.Anything, id, mock.Anything).Return(nil, grpc.Errorf(codes.Internal, "boom"))

	_, err := e.Create(context.Background(), id)

	assert.Error(t, err)
	assert.Equal(t, codes.Internal, grpc.Code(err))
}

func TestCreateDoesNotAddItemToTheQueueIfInCache(t *testing.T) {
	e := setup(t, 0, 0)
	id := &wrappers.StringValue{Value: "abc"}
	mockCache.ExpectedCalls = make([]*mock.Call, 0)
	mockCache.On("Exists", mock.Anything, id, mock.Anything).Return(&wrappers.BoolValue{Value: true}, nil)

	i, err := e.Create(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	mockQueue.AssertNotCalled(t, "Position", mock.Anything)
	mockQueue.AssertNotCalled(t, "Push", mock.Anything)
	assert.Equal(t, id.GetValue(), i.Id)
	assert.Equal(t, &emojify.QueryStatus{Status: emojify.QueryStatus_FINISHED}, i.GetStatus())
}

func TestCreateDoesNotAddItemToTheQueueIfPresent(t *testing.T) {
	e := setup(t, 1, 2)
	id := &wrappers.StringValue{Value: "abc"}

	i, err := e.Create(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	mockQueue.AssertCalled(t, "Position", mock.Anything)
	mockQueue.AssertNotCalled(t, "Push", mock.Anything)
	assert.Equal(t, id.GetValue(), i.Id)
	assert.Equal(t, int32(1), i.QueuePosition)
	assert.Equal(t, int32(2), i.QueueLength)
	assert.Equal(t, &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED}, i.GetStatus())
}
