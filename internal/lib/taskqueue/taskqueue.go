package taskqueue

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

type TaskQueue struct {
	redis *redis.Client

	queueName   string
	maxAttempts int

	pendingKey    string
	processingKey string
	completedKey  string
	failedKey     string
	taskPrefix    string
}

type Options struct {
	queueName   string
	maxAttempts int
}

func New(opts Options) *TaskQueue {
	if opts.queueName == "" {
		opts.queueName = "tasks"
	}
	if opts.maxAttempts <= 0 {
		opts.maxAttempts = 3
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Protocol: 2,
	})

	return &TaskQueue{
		redis: rdb,

		queueName:   opts.queueName,
		maxAttempts: opts.maxAttempts,

		pendingKey:    fmt.Sprintf("queue:%s:pending", opts.queueName),
		processingKey: fmt.Sprintf("queue:%s:processing", opts.queueName),
		completedKey:  fmt.Sprintf("queue:%s:completed", opts.queueName),
		failedKey:     fmt.Sprintf("queue:%s:failed", opts.queueName),
		taskPrefix:    fmt.Sprintf("queue:%s:task:", opts.queueName),
	}
}
