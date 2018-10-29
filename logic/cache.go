package logic

import (
	"context"
	"fmt"
	"log"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/golang/protobuf/ptypes/wrappers"
	resolver "github.com/nicholasjackson/grpc-consul-resolver"
	"google.golang.org/grpc"
)

// Cache defines an interface for an image cache
type Cache interface {
	// Exists checks if an item exists in the cache
	Exists(string) (bool, error)
	// Get an image from the cache, returns true, image if found in cache or false, nil if image not found
	Get(string) ([]byte, error)
	// Put an image into the cache, returns an error if unsuccessful
	Put(string, []byte) error
}

// RemoteCache implements Cache to interact wtih the remote cachee
type RemoteCache struct {
	client cache.CacheClient
}

// NewCache creates a new RemoteCache
func NewCache(cacheService, consulAddress, serviceName string) (Cache, error) {
	log.Println("Creating cache for", cacheService)
	r, dialer, _ := resolver.NewConnectServiceQueryResolver(consulAddress, serviceName)

	lb := grpc.RoundRobin(r)

	fmt.Println(cacheService)
	// create a new gRPC client connection
	c, err := grpc.Dial(
		cacheService,
		grpc.WithInsecure(),
		grpc.WithBalancer(lb),
		grpc.WithBlock(),
		dialer,
	)

	if err != nil {
		return nil, err
	}

	cc := cache.NewCacheClient(c)
	rc := &RemoteCache{cc}

	return rc, nil
}

// Exists checks if an item exists in the cache
func (r *RemoteCache) Exists(uri string) (bool, error) {
	exists, err := r.client.Exists(context.Background(), &wrappers.StringValue{Value: uri})
	if err != nil {
		return false, err
	}

	return exists.Value, err
}

// Get an image from the cache, returns true, image if found in cache or false, nil if image not found
func (r *RemoteCache) Get(uri string) ([]byte, error) {
	data, err := r.client.Get(context.Background(), &wrappers.StringValue{Value: uri})
	return data.GetData(), err
}

// Put an image into the cache, returns an error if unsuccessful
func (r *RemoteCache) Put(uri string, data []byte) error {
	ci := cache.CacheItem{
		Id:   uri,
		Data: data,
	}

	_, err := r.client.Put(context.Background(), &ci)
	return err
}
