package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/hashicorp/go-hclog"
)

// Redis is a queue implementation for the Redis server
type Redis struct {
	client      *redis.Client
	list        string
	expiration  time.Duration
	popChan     chan PopResponse
	doneChan    chan PopResponse
	currentItem *Item
	logger      hclog.Logger
	errorDelay  time.Duration
}

// New creates a new Redis queue
func New(addr, password string, db int, l hclog.Logger) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &Redis{
		client:     client,
		list:       "worker_queue",
		expiration: 30 * time.Minute,
		logger:     l,
		errorDelay: 5 * time.Second,
		popChan:    make(chan PopResponse),
		doneChan:   make(chan PopResponse),
	}, nil
}

// Push an item onto the queue
func (r *Redis) Push(i *Item) (position int, length int, err error) {
	//serialize the item to json
	j, err := json.Marshal(i)
	if err != nil {
		return 0, 0, fmt.Errorf("unable marshal item to json: %s", err)
	}

	// add the item to the db
	s := r.client.Set(i.ID, string(j), r.expiration)
	if err := s.Err(); err != nil {
		return 0, 0, fmt.Errorf("unable to add item to set: %s", err)
	}

	// store the item in a ordered set
	c := r.client.ZAdd(r.list, redis.Z{Score: float64(time.Now().UnixNano()), Member: i.ID})
	if err := c.Err(); err != nil {
		return 0, 0, fmt.Errorf("unable to add item to ordered list: %s", err)
	}

	return r.Position(i.ID)
}

// Pop returns a channel containing items from the front of the queue
func (r *Redis) Pop() chan PopResponse {
	go func() {
		// loop over the queue constantly returning items
		for {
			// get the first key from the set
			k := r.client.ZPopMin(r.list, 1)
			if err := k.Err(); err != nil {
				r.logger.Error("Error reading from queue", "error", err)

				time.Sleep(r.errorDelay)
				continue
			}

			res, err := k.Result()
			if err != nil {
				r.logger.Error("Error getting result from queue item", "error", err)

				time.Sleep(r.errorDelay)
				continue
			}

			// check that an item has been returned, if not sleep
			if len(res) < 1 {
				r.logger.Trace("No items in queue item", "error", err)

				time.Sleep(r.errorDelay)
				continue
			}

			// get the corresponding item from the db
			key := res[0].Member
			i := r.client.Get(key.(string))
			if err := i.Err(); err != nil {
				r.logger.Error("Queue item not in database", "error", err)

				time.Sleep(r.errorDelay)
				continue
			}

			// delete the item from the db now it has been retrieved
			r.client.Del(key.(string))

			// unmarshal the item
			data, err := i.Result()
			if err != nil {
				r.logger.Error("Deleting queue item from database", "error", err)

				time.Sleep(r.errorDelay)
				continue
			}

			item := &Item{}
			err = json.Unmarshal([]byte(data), item)
			if err != nil {
				r.logger.Error("Unable to marshal item from database", "error", err)

				time.Sleep(r.errorDelay)
				continue
			}

			r.logger.Debug("Send item from queue to worker", "item", item)

			// store the currently processing item in case the client
			// queries as it has now been removed from the queue
			r.currentItem = item

			// block until a worker is able to accept the request
			r.popChan <- PopResponse{Item: item, Done: r.doneChan}

			r.logger.Debug("Waiting for worker to complete", "item", item)

			// block until the worker has processed the item
			select {
			case pr := <-r.doneChan:
				r.currentItem = nil

				if pr.Error != nil {
					// TODO handle requeueing failed items
					r.logger.Error("Item processing failed", "item", pr.Item, "error", pr.Error)
				} else {
					r.logger.Debug("Item processing complete queue", "item", pr.Item)
				}
			}

		}
	}()

	return r.popChan
}

// Position allows you to query the position of an item in the queue
func (r *Redis) Position(key string) (position, length int, err error) {
	max := r.client.ZCount(r.list, "-inf", "+inf")
	if err := max.Err(); err != nil {
		return 0, 0, fmt.Errorf("unable to get set count: %s", err)
	}

	if r.currentItem != nil {
		r.logger.Debug("Current item", "item", r.currentItem.ID, "key", key)

		// if the key is the current item id then return the item as we are processing
		if key == r.currentItem.ID {
			return -1, int(max.Val() + 1), nil
		}
	}

	// if the queue is empty do not lookup
	if max.Val() == 0 {
		return 0, 0, nil
	}

	// otherwise return the item from the list
	pos := r.client.ZRank(r.list, key)
	if err := pos.Err(); err != nil {
		return 0, 0, fmt.Errorf("unable to find item position: %s", err)
	}

	return int(pos.Val() + 1), int(max.Val()), nil
}

// Ping Redis to check up
func (r *Redis) Ping() error {
	status := r.client.Ping()
	if err := status.Err(); err != nil {
		return err
	}

	return nil
}
