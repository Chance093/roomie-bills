package taskqueue

import (
	"context"
	"fmt"
	"sync"
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

	statsMu    sync.Mutex
	enqueuedN  int
	completedN int
	failedN    int
	reclaimedN int
}

type Options struct {
	queueName        string
	maxAttempts      int
	completedTTL     int
	completedHistory int
	reclaimMs        int
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
	if opts.reclaimMs <= 0 {
		opts.reclaimMs = 10000 // 10 seconds
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
			reclaimMs:        opts.reclaimMs,
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

	q.statsMu.Lock()
	q.enqueuedN++
	q.statsMu.Unlock()

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
		return false, fmt.Errorf("Error while performing pipeline redis calls for completed task (%s): %w", taskId, err)
	}

	q.statsMu.Lock()
	q.completedN++
	q.statsMu.Unlock()

	return true, nil
}

func (q *taskQueue) Fail(ctx context.Context, task *ClaimedTask, errMsg error) (bool, error) {
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

	// get task hash from redis
	var t Task
	if err := q.redis.HGetAll(ctx, taskKey).Scan(&t); err != nil {
		return false, fmt.Errorf("Error while getting task (%s) hash from redis: %w", taskId, err)
	}

	// remove task id from processing queue, update task hash properties based on
	// whether or not the task should retry, and if the task should retry, it should
	// push the task id to the pending queue, else it should push to the failed queue,
	// trim the failed queue, and set expiration on task hash
	pipe := q.redis.Pipeline()
	pipe.LRem(ctx, q.processingKey, 1, taskId)

	retry := task.Attempts < q.maxAttempts
	if retry {
		t.Status = "pending"
		t.LastError = errMsg.Error()
		t.LastErrorAtMs = nowMs()
		t.ClaimToken = ""
		t.ClaimedAtMs = 0
		pipe.LPush(ctx, q.pendingKey, taskId)
	} else {
		t.Status = "failed"
		t.LastError = errMsg.Error()
		t.LastErrorAtMs = nowMs()
		t.ClaimToken = ""
		pipe.LPush(ctx, q.failedKey, taskId)
		pipe.LTrim(ctx, q.failedKey, 0, int64(q.completedHistory)-1)
		pipe.Expire(ctx, taskKey, time.Duration(q.completedTTL))
	}

	pipe.HSet(ctx, taskKey, t)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("Error while performing pipeline redis calls for failed task (%s): %w", taskId, err)
	}

	q.statsMu.Lock()
	q.failedN++
	q.statsMu.Unlock()

	return true, nil
}

// go through processing list and find stuck jobs
// if stuck, put back in pending
func (q *taskQueue) ReclaimStuck(ctx context.Context) ([]string, error) {
	processing, err := q.redis.LRange(ctx, q.processingKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("Error while getting all tasks in processing queue: %w", err)
	}
	var reclaimed []string

	for _, taskId := range processing {
		// get task hash from redis
		taskKey := q.taskKey(taskId)
		var t Task
		if err := q.redis.HGetAll(ctx, taskKey).Scan(&t); err != nil {
			return nil, fmt.Errorf("Error while getting task (%s) hash from redis: %w", taskId, err)
		}

		// A job is past the timeout if either:
		//   - claimed_at_ms is set and (now - claimed_at_ms) > reclaimMs, OR
		//   - claimed_at_ms is missing (worker crashed between BLMOVE and the
		//     metadata write) and (now - enqueued_at_ms) > 2 * reclaimMs.
		now := nowMs()
		if (t.ClaimedAtMs > 0 && (now-t.ClaimedAtMs) > int64(q.reclaimMs)) ||
			(t.ClaimedAtMs == 0 && (now-t.EnqueuedAtMs) > 2*int64(q.reclaimMs)) {
			t.Status = "pending"
			t.ReclaimedAtMs = now
			t.ClaimedAtMs = 0
			t.ClaimToken = ""

			pipe := q.redis.Pipeline()
			pipe.LRem(ctx, q.processingKey, 1, taskId)
			pipe.LPush(ctx, q.pendingKey, taskId)
			pipe.HSet(ctx, taskKey, t)
			if _, err := pipe.Exec(ctx); err != nil {
				return nil, fmt.Errorf("Error while performing pipeline redis calls for stuck task (%s): %w", taskId, err)
			}

			reclaimed = append(reclaimed, taskId)
		}
	}

	q.statsMu.Lock()
	q.reclaimedN += len(reclaimed)
	q.statsMu.Unlock()

	return reclaimed, nil
}

type QueueStats struct {
	EnqueuedTotal   int
	CompletedTotal  int
	FailedTotal     int
	ReclaimedTotal  int
	PendingDepth    int64
	ProcessingDepth int64
	CompletedDepth  int64
	FailedDepth     int64
}

func (q *taskQueue) Stats(ctx context.Context) (QueueStats, error) {
	pipe := q.redis.Pipeline()
	pendingCmd := pipe.LLen(ctx, q.pendingKey)
	processingCmd := pipe.LLen(ctx, q.processingKey)
	completedCmd := pipe.LLen(ctx, q.completedKey)
	failedCmd := pipe.LLen(ctx, q.failedKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return QueueStats{}, fmt.Errorf("Error while performing pipeline redis calls for stats: %w", err)
	}

	q.statsMu.Lock()
	defer q.statsMu.Unlock()

	return QueueStats{
		EnqueuedTotal:   q.enqueuedN,
		CompletedTotal:  q.completedN,
		FailedTotal:     q.failedN,
		ReclaimedTotal:  q.reclaimedN,
		PendingDepth:    pendingCmd.Val(),
		ProcessingDepth: processingCmd.Val(),
		CompletedDepth:  completedCmd.Val(),
		FailedDepth:     failedCmd.Val(),
	}, nil
}

func nowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (q *taskQueue) taskKey(taskId string) string {
	return q.taskPrefix + taskId
}
