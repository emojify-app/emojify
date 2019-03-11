package emojify

import (
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/mock"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// MockClient is a mock implementation of the gRPC client for testing
type MockClient struct {
	mock.Mock
}

// Check is a mock implementation of the Check interface method
func (m *MockClient) Check(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	args := m.Called(ctx, in, opts)

	if hr := args.Get(0); hr != nil {
		return hr.(*HealthCheckResponse), nil
	}

	return nil, args.Error(1)
}

// Create is a mock implementation of the Create interface method
func (m *MockClient) Create(ctx context.Context, in *wrappers.StringValue, opts ...grpc.CallOption) (*QueryItem, error) {
	args := m.Called(ctx, in, opts)

	if qi := args.Get(0); qi != nil {
		return qi.(*QueryItem), args.Error(1)
	}

	return nil, args.Error(1)
}

// Query is a mock implementation of the Query interface method
func (m *MockClient) Query(ctx context.Context, in *wrappers.StringValue, opts ...grpc.CallOption) (*QueryItem, error) {
	args := m.Called(ctx, in, opts)

	if qi := args.Get(0); qi != nil {
		return qi.(*QueryItem), args.Error(1)
	}

	return nil, args.Error(1)
}
