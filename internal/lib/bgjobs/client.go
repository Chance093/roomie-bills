package bgjobs

import (
	"context"
	"encoding/json"
	"fmt"
)

type Client struct {
	primeQ queue
}

func NewClient(opts RedisOpts) Client {
	rdb := initNewRDBClient(opts)
	pq := newPrimaryQueue(rdb)

	c := Client{pq}

	return c
}

func (c Client) Close() { c.primeQ.rdb.Close() }

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
	ctx := context.TODO()
	if err := c.primeQ.enqueue(ctx, jTask); err != nil {
		return TaskInfo{}, fmt.Errorf("Failed to enqueue task: %w", err)
	}

	return TaskInfo{Id: task.Id, Queue: c.primeQ.name}, nil
}
