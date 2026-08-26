package bgjobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisOpts struct {
	Addr string
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

func (c Client) Enqueue(task *Task, opts ...any) (TaskInfo, error) {
	// override task opts with new opts
	setTaskOpts(task, opts)

	// marshall json
	jTask, err := json.Marshal(*task)
	if err != nil {
		return TaskInfo{}, fmt.Errorf("Failed to marshal task into json: %w", err)
	}

	// push to redis message queue
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})
	ctx := context.Background()

	if _, err := rdb.LPush(ctx, Queue, jTask).Result(); err != nil {
		return TaskInfo{}, fmt.Errorf("Failed to enqueue task: %w", err)
	}

	return TaskInfo{Id: task.Id, Queue: Queue}, nil
}
