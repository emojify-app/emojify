package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// Redis is a queue implementation for the Redis server
type Redis struct {
	client     *redis.Client
	list       string
	expiration time.Duration
	popChan    chan PopResponse
}

// New creates a new Redis queue
func New(addr, password string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &Redis{client, "worker_queue", 30 * time.Minute, make(chan PopResponse)}, nil
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
		for {
			// get the first key from the set
			k := r.client.ZPopMin(r.list, 1)
			if err := k.Err(); err != nil {
				r.popChan <- PopResponse{Error: err}
				continue
			}

			res, err := k.Result()
			if err != nil {
				r.popChan <- PopResponse{Error: err}
				continue
			}

			// check that an item has been returned, if not sleep
			if len(res) < 1 {
				r.popChan <- PopResponse{}
				continue
			}

			// get the corresponding item from the db
			key := res[0].Member
			i := r.client.Get(key.(string))
			if err := i.Err(); err != nil {
				r.popChan <- PopResponse{Error: err}
				continue
			}

			// delete the item from the db now it has been retrieved
			r.client.Del(key.(string))

			// unmarshal the item
			item := &Item{}
			err = json.Unmarshal([]byte(i.String()), item)
			if err != nil {
				r.popChan <- PopResponse{Error: err}
				continue
			}

			r.popChan <- PopResponse{Item: item}
		}
	}()

	return r.popChan
}

// Position allows you to query the position of an item in the queue
func (r *Redis) Position(key string) (position, length int, err error) {
	pos := r.client.ZRank(r.list, key)
	if err := pos.Err(); err != nil {
		return 0, 0, fmt.Errorf("unable to find item position: %s", err)
	}

	max := r.client.ZCount(r.list, "-inf", "+inf")
	if err := max.Err(); err != nil {
		return 0, 0, fmt.Errorf("unable to get set count: %s", err)
	}

	return int(pos.Val()), int(max.Val()), nil
}
