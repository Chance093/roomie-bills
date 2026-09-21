package taskqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type taskQueue struct {
	redis *redis.Client

	Options

	pendingKey    string
	processingKey string
	completedKey  string
	failedKey     string
	taskPrefix    string
}

type Options struct {
	queueName        string
	maxAttempts      int
	completedTTL     int
	completedHistory int
}

func newTaskQueue(opts Options) *taskQueue {
	if opts.queueName == "" {
		opts.queueName = "tasks"
	}
	if opts.maxAttempts <= 0 {
		opts.maxAttempts = 3
	}
	if opts.completedTTL <= 0 {
		opts.completedTTL = 300
	}
	if opts.completedHistory <= 0 {
		opts.completedHistory = 50
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Protocol: 2,
	})

	return &taskQueue{
		redis: rdb,

		Options: Options{
			queueName:        opts.queueName,
			maxAttempts:      opts.maxAttempts,
			completedTTL:     opts.completedTTL,
			completedHistory: opts.completedHistory,
		},

		pendingKey:    fmt.Sprintf("queue:%s:pending", opts.queueName),
		processingKey: fmt.Sprintf("queue:%s:processing", opts.queueName),
		completedKey:  fmt.Sprintf("queue:%s:completed", opts.queueName),
		failedKey:     fmt.Sprintf("queue:%s:failed", opts.queueName),
		taskPrefix:    fmt.Sprintf("queue:%s:task:", opts.queueName),
	}
}

func (q *taskQueue) Enqueue(ctx context.Context, t Task) (string, error) {
	// update task properties
	taskId := t.Id
	t.EnqueuedAtMs = nowMs()
	t.Status = "pending"

	// set task hash and push to pending task queue in redis
	pipe := q.redis.Pipeline()
	pipe.HSet(ctx, q.taskKey(taskId), t)
	pipe.LPush(ctx, q.pendingKey, taskId)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("Error while enqueueing task in redis: %w", err)
	}

	return taskId, nil
}

type ClaimedTask struct {
	Id         string
	Payload    string
	Attempts   int
	ClaimToken string
}

func (q *taskQueue) Claim(ctx context.Context) (*ClaimedTask, error) {
	// move task id from pending task queue to processing task queue
	taskId, err := q.redis.BLMove(ctx, q.pendingKey, q.processingKey, "RIGHT", "LEFT", time.Duration(0)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Error while moving taskId from pending to processing queue: %w", err)
	}

	// get task hash from redis and update task properties
	var t Task
	taskKey := q.taskKey(taskId)
	if err := q.redis.HGetAll(ctx, taskKey).Scan(&t); err != nil {
		return nil, fmt.Errorf("Error while getting task (%s) hash from redis: %w", taskId, err)
	}

	t.Status = "processing"
	t.ClaimedAtMs = nowMs()
	t.ClaimToken = uuid.New().String()
	t.Attempts++

	// update task hash in redis and return claimed task
	if _, err := q.redis.HSet(ctx, taskKey, t).Result(); err != nil {
		return nil, fmt.Errorf("Error while updating task (%s) hash in redis: %w", taskId, err)
	}

	return &ClaimedTask{taskId, t.Payload, t.Attempts, t.ClaimToken}, nil
}

func (q *taskQueue) Complete(ctx context.Context, task *ClaimedTask) (bool, error) {
	// compare claim token in claimed task with claim token in redis
	taskId := task.Id
	taskKey := q.taskKey(taskId)
	currentToken, err := q.redis.HGet(ctx, taskKey, "claim_token").Result()
	if err != nil {
		return false, fmt.Errorf("Error while getting claim token from task (%s) hash: %w", taskId, err)
	}

	if currentToken != task.ClaimToken {
		return false, nil // TODO: error handling (don't know if it should be error)
	}

	// get task hash from redis and update task properties
	var t Task
	if err := q.redis.HGetAll(ctx, taskKey).Scan(&t); err != nil {
		return false, fmt.Errorf("Error while getting task (%s) hash from redis: %w", taskId, err)
	}

	t.Status = "completed"
	t.CompletedAtMs = nowMs()

	// remove task id from processing task queue, update task hash, set expiration 
	// for task hash, push task id to completed task queue, and trim the completed 
	// task queue, and execute these commands all at once
	pipe := q.redis.Pipeline()
	pipe.LRem(ctx, q.processingKey, 1, taskId)
	pipe.HSet(ctx, taskKey, t)
	pipe.Expire(ctx, taskKey, time.Duration(q.completedTTL))
	pipe.LPush(ctx, q.completedKey, taskId)
	pipe.LTrim(ctx, q.completedKey, 0, int64(q.completedHistory)-1)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("Error while performing pipeline redis calls for completed task: %w", err)
	}

	return true, nil
}

func nowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (q taskQueue) taskKey(taskId string) string {
	return q.taskPrefix + taskId
}
