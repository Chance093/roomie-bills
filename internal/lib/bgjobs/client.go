package bgjobs

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisOpts struct {
	addr string
}

type Client struct {
	RedisOpts
}

func NewClient(opts RedisOpts) Client {
	c := Client{opts}

	return c
}

type TaskInfo struct {
	Id    string
	Queue string
}

// TODO: error handling
func (c Client) Enqueue(task Task) (TaskInfo, error) {
	const Queue = "JobQueue"
	// marshall json
	jTask, err := json.Marshal(task)
	if err != nil {
		return TaskInfo{}, err
	}

	// TODO: Generate example id
	id := "example id"

	// push to redis message queue
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})
	ctx := context.Background()

	if _, err := rdb.LPush(ctx, Queue, jTask).Result(); err != nil {
		return TaskInfo{}, err
	}

	return TaskInfo{Id: id, Queue: Queue}, nil
}
