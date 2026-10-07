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

	TaskQueueOpts

	// Redis key for pending queue.
	pendingKey string
	// Redis key for processing queue.
	processingKey string
	// Redis key for completed queue.
	completedKey string
	// Redis key for failed queue.
	failedKey string
	// Prefix for all TaskMeta hash keys.
	// Format: "queue:{queueName}:tasks:{taskId}"
	taskPrefix string

	// Mutex to atomically increment stats
	statsMu sync.Mutex
	// Amount of tasks that have been enqueued
	enqueuedN int
	// Amount of tasks that have been completed
	completedN int
	// Amount of tasks that have been failed
	failedN int
	// Amount of tasks that have been reclaimed
	reclaimedN int
}

type TaskQueueOpts struct {
	// Formatted in redis as "queue:{queueName}:{queueType}".
	// Default: "tasks" ("queue:tasks:{queueType}").
	QueueName string

	// Max attempts a worker will try to process a task.
	// Default: 3.
	MaxAttempts int

	// Time-To-Live (seconds) for completed tasks.
	// Default: 300.
	CompletedTTL int

	// The max amount of tasks allowed to sit in completed queue before trim.
	// Default: 50.
	CompletedHistory int

	// The amount of time (ms) a stuck task should wait before being reclaimed.
	// Default: 10,000 (10 seconds).
	ReclaimMs int

	RedisClientOpts
}

type RedisClientOpts struct {
	// Addr is the address formated as host:port
	Addr string

	// Username is used to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string

	// Password is an optional password. Must match the password specified in the
	// `requirepass` server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string
}

// Initializes a redis client within a task queue to be passed as
// a custom queue to either NewServerWithCustomQueue() or
// NewClientWithCustomQueue().
func NewTaskQueue(opts TaskQueueOpts) *taskQueue {
	if opts.RedisClientOpts.Addr == "" {
		opts.RedisClientOpts.Addr = "127.0.0.1:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     opts.RedisClientOpts.Addr,
		Username: opts.RedisClientOpts.Username,
		Password: opts.RedisClientOpts.Password,
	})

	if opts.QueueName == "" {
		opts.QueueName = "tasks"
	}

	q := &taskQueue{
		redis: rdb,

		TaskQueueOpts: opts,

		pendingKey:    fmt.Sprintf("queue:%s:pending", opts.QueueName),
		processingKey: fmt.Sprintf("queue:%s:processing", opts.QueueName),
		completedKey:  fmt.Sprintf("queue:%s:completed", opts.QueueName),
		failedKey:     fmt.Sprintf("queue:%s:failed", opts.QueueName),
		taskPrefix:    fmt.Sprintf("queue:%s:task:", opts.QueueName),
	}

	q.setDefaultOpts()

	return q
}

// sets default opts for the task queue
func (q *taskQueue) setDefaultOpts() {
	if q.MaxAttempts <= 0 {
		q.MaxAttempts = 3
	}
	if q.CompletedTTL <= 0 {
		q.CompletedTTL = 300
	}
	if q.CompletedHistory <= 0 {
		q.CompletedHistory = 50
	}
	if q.ReclaimMs <= 0 {
		q.ReclaimMs = 10000 // 10 seconds
	}
}

// Allows a client to enqueue a task. Stores the TaskMeta as a hash, and
// then pushes the TaskMeta id to a pending queue. Also increments the
// enqueuedN property atomically.
func (q *taskQueue) enqueue(ctx context.Context, t TaskMeta) (string, error) {
	// update task properties
	taskId := t.Id
	t.EnqueuedAtMs = nowMs()
	t.Status = "pending"

	// set task hash and push to pending task queue in redis
	pipe := q.redis.TxPipeline()
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

// Allows a worker to claim a task. Moves a TaskMeta id from the pending
// queue to a processing queue. It then updates some TaskMeta properties
// to reflect that the Task is processing and is claimed.
func (q *taskQueue) claim(ctx context.Context, timeoutMs int) (*claimedTask, error) {
	// set timeout for blocking move
	timeout := max(time.Duration(timeoutMs)*time.Millisecond, 100*time.Millisecond)

	// move task id from pending task queue to processing task queue
	taskId, err := q.redis.BLMove(ctx, q.pendingKey, q.processingKey, "RIGHT", "LEFT", timeout).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Error while moving taskId from pending to processing queue: %w", err)
	}

	// get task hash from redis and update task properties
	var t TaskMeta
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

	return &claimedTask{taskId, t.Name, t.Payload, t.TimeoutMs, t.Attempts, t.ClaimToken}, nil
}

// Allows a worker to complete a task. Moves a TaskMeta id from the
// processing queue to a completed queue. It then updates TaskMeta properties
// to show completed. Also increments the completedN property atomically.
func (q *taskQueue) complete(ctx context.Context, task *claimedTask) (bool, error) {
	// compare claim token in claimed task with claim token in redis
	taskId := task.Id
	taskKey := q.taskKey(taskId)
	currentToken, err := q.redis.HGet(ctx, taskKey, "claim_token").Result()
	if err != nil {
		return false, fmt.Errorf("Error while getting claim token from task (%s) hash: %w", taskId, err)
	}

	if currentToken != task.ClaimToken { // this job was reclaimed and og claimer is trying to complete
		return false, fmt.Errorf("Task was reclaimed and OG claimer tried to complete task. OG claim token: %s - Current claim token: %s", task.ClaimToken, currentToken)
	}

	// get task hash from redis and update task properties
	var t TaskMeta
	if err := q.redis.HGetAll(ctx, taskKey).Scan(&t); err != nil {
		return false, fmt.Errorf("Error while getting task (%s) hash from redis: %w", taskId, err)
	}

	t.Status = "completed"
	t.CompletedAtMs = nowMs()

	// remove task id from processing task queue, update task hash, set expiration
	// for task hash, push task id to completed task queue, and trim the completed
	// task queue, and execute these commands all at once
	pipe := q.redis.TxPipeline()
	pipe.LRem(ctx, q.processingKey, 1, taskId)
	pipe.HSet(ctx, taskKey, t)
	pipe.Expire(ctx, taskKey, time.Duration(q.CompletedTTL))
	pipe.LPush(ctx, q.completedKey, taskId)
	pipe.LTrim(ctx, q.completedKey, 0, int64(q.CompletedHistory)-1)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("Error while performing pipeline redis calls for completed task (%s): %w", taskId, err)
	}

	q.statsMu.Lock()
	q.completedN++
	q.statsMu.Unlock()

	return true, nil
}

// Allows a worker to fail a task. Moves a TaskMeta id from the processing
// queue to a failed queue. If the task has not reached max attempts, it moves
// the task back to pending instead. It then updates TaskMeta properties to
// show failed or pending. Also increments the failedN property atomically.
func (q *taskQueue) fail(ctx context.Context, task *claimedTask, errMsg error) (bool, error) {
	// compare claim token in claimed task with claim token in redis
	taskId := task.Id
	taskKey := q.taskKey(taskId)
	currentToken, err := q.redis.HGet(ctx, taskKey, "claim_token").Result()
	if err != nil {
		return false, fmt.Errorf("Error while getting claim token from task (%s) hash: %w", taskId, err)
	}
	if currentToken != task.ClaimToken { // this job was reclaimed and og claimer is trying to fail
		return false, fmt.Errorf("Task was reclaimed and OG claimer tried to fail task. OG claim token: %s - Current claim token: %s", task.ClaimToken, currentToken)
	}

	// get task hash from redis
	var t TaskMeta
	if err := q.redis.HGetAll(ctx, taskKey).Scan(&t); err != nil {
		return false, fmt.Errorf("Error while getting task (%s) hash from redis: %w", taskId, err)
	}

	// remove task id from processing queue, update task hash properties based on
	// whether or not the task should retry, and if the task should retry, it should
	// push the task id to the pending queue, else it should push to the failed queue,
	// trim the failed queue, and set expiration on task hash
	pipe := q.redis.TxPipeline()
	pipe.LRem(ctx, q.processingKey, 1, taskId)

	retry := task.Attempts < q.MaxAttempts
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
		pipe.LTrim(ctx, q.failedKey, 0, int64(q.CompletedHistory)-1)
		pipe.Expire(ctx, taskKey, time.Duration(q.CompletedTTL))
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

// When a worker gets stuck (crashes between a BLMOVE and metadata write or
// server crashes while processing), will move stuck tasks back to pending queue.
// Also increments the reclaimedN property atomically.
func (q *taskQueue) reclaimStuck(ctx context.Context) ([]string, error) {
	processing, err := q.redis.LRange(ctx, q.processingKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("Error while getting all tasks in processing queue: %w", err)
	}
	var reclaimed []string

	for _, taskId := range processing {
		// get task hash from redis
		taskKey := q.taskKey(taskId)
		var t TaskMeta
		if err := q.redis.HGetAll(ctx, taskKey).Scan(&t); err != nil {
			return nil, fmt.Errorf("Error while getting task (%s) hash from redis: %w", taskId, err)
		}

		// A job is past the timeout if either:
		//   - claimed_at_ms is set and (now - claimed_at_ms) > reclaimMs, OR
		//   - claimed_at_ms is missing (worker crashed between BLMOVE and the
		//     metadata write) and (now - enqueued_at_ms) > 2 * reclaimMs.
		now := nowMs()
		if (t.ClaimedAtMs > 0 && (now-t.ClaimedAtMs) > int64(q.ReclaimMs)) ||
			(t.ClaimedAtMs == 0 && (now-t.EnqueuedAtMs) > 2*int64(q.ReclaimMs)) {
			t.Status = "pending"
			t.ReclaimedAtMs = now
			t.ClaimedAtMs = 0
			t.ClaimToken = ""

			pipe := q.redis.TxPipeline()
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

type queueStats struct {
	// Amount of tasks that have been enqueued
	enqueuedTotal int
	// Amount of tasks that have been completed
	completedTotal int
	// Amount of tasks that have been failed
	failedTotal int
	// Amount of tasks that have been reclaimed
	reclaimedTotal int
	// Depth of the pending queue
	pendingDepth int64
	// Depth of the processing queue
	processingDepth int64
	// Depth of the completed queue
	completedDepth int64
	// Depth of the failed queue
	failedDepth int64
}

// Returns the stats associated with the queue such as the depth of all queues,
// and the total of all task statuses.
func (q *taskQueue) stats(ctx context.Context) (queueStats, error) {
	pipe := q.redis.TxPipeline()
	pendingCmd := pipe.LLen(ctx, q.pendingKey)
	processingCmd := pipe.LLen(ctx, q.processingKey)
	completedCmd := pipe.LLen(ctx, q.completedKey)
	failedCmd := pipe.LLen(ctx, q.failedKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return queueStats{}, fmt.Errorf("Error while performing pipeline redis calls for stats: %w", err)
	}

	q.statsMu.Lock()
	defer q.statsMu.Unlock()

	return queueStats{
		enqueuedTotal:   q.enqueuedN,
		completedTotal:  q.completedN,
		failedTotal:     q.failedN,
		reclaimedTotal:  q.reclaimedN,
		pendingDepth:    pendingCmd.Val(),
		processingDepth: processingCmd.Val(),
		completedDepth:  completedCmd.Val(),
		failedDepth:     failedCmd.Val(),
	}, nil
}

// Resets the stats of the task queue back to 0.
func (q *taskQueue) resetStats() {
	q.statsMu.Lock()
	defer q.statsMu.Unlock()

	q.enqueuedN = 0
	q.completedN = 0
	q.failedN = 0
	q.reclaimedN = 0
}

// Purges all data from all the queues and TaskMeta hashes. Also resets
// the task queue stats.
func (q *taskQueue) purge(ctx context.Context) error {
	pipe := q.redis.Pipeline()
	pipe.Del(ctx, q.pendingKey, q.processingKey, q.completedKey, q.failedKey)

	var cursor uint64
	pattern := q.taskPrefix + "*"

	for {
		keys, next, err := q.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		for _, k := range keys {
			pipe.Del(ctx, k)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	q.resetStats()
	return nil
}

// Returns the current time in milliseconds.
func nowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Formats a task id into a task key for a redis hash key.
// Example: 123456 -> queue:{queueName}:task:123456
func (q *taskQueue) taskKey(taskId string) string {
	return q.taskPrefix + taskId
}
