package taskqueue

import (
	"github.com/google/uuid"
)

type Task struct {
	// Name of the task (maps to task handlers).
	Name string

	// Payload of task stringified (for redis hash storage).
	Payload string

	TaskOpts
}

type TaskOpts struct {
	// Amount of time (ms) before a task times out and fails.
	TimeoutMs int
}

// Returns a new task meant to be enqueued by Client.
func NewTask(name string, payload []byte, opts TaskOpts) Task {
	t := Task{
		Name:     name,
		Payload:  string(payload),
		TaskOpts: TaskOpts{opts.TimeoutMs},
	}

	t.setDefaultOpts()

	return t
}

// sets default opts for task
func (t *Task) setDefaultOpts() {
	if t.TimeoutMs <= 0 {
		t.TimeoutMs = 5000 // 5 seconds
	}
}

type TaskMeta struct {
	// Random uuid
	Id string `redis:"id"`

	// Name of the task (maps to task handlers).
	Name string `redis:"name"`

	// Payload of task stringified (for redis hash storage).
	Payload string `redis:"payload"`

	// Amount of time (ms) before a task times out and fails.
	TimeoutMs int `redis:"timeout_ms"`

	// "Pending" | "Processing" | "Completed" | "Failed"
	Status string `redis:"status"`

	// Amount of attempts a worker has tried to process the task
	Attempts int `redis:"attempts"`

	// Time in ms that the task was enqueued
	EnqueuedAtMs int64 `redis:"enqueued_at_ms"`

	// Time in ms that the task was claimed
	ClaimedAtMs int64 `redis:"claimed_at_ms"`

	// Time in ms that the task was completed
	CompletedAtMs int64 `redis:"completed_at_ms"`

	// Time in ms that a stuck task was reclaimed
	ReclaimedAtMs int64 `redis:"reclaimed_at_ms"`

	// Unique token that helps identify which worker has claim of task
	ClaimToken string `redis:"claim_token"`

	// Last known error that was returned after processing a task
	LastError string `redis:"last_error"`

	// Time in ms that a task failed
	LastErrorAtMs int64 `redis:"last_error_at_ms"`
}

// Returns new TaskMeta, which is only to be used for storing as a
// hash in redis.
func newTaskMeta(task Task) TaskMeta {
	return TaskMeta{
		Id:        uuid.New().String(),
		Name:      task.Name,
		Payload:   task.Payload,
		TimeoutMs: task.TimeoutMs,
	}
}

type ClaimedTask struct {
	// Random uuid
	Id string

	// Name of the task (maps to task handlers).
	Name string

	// Payload of task stringified (for redis hash storage).
	Payload string

	// Amount of time (ms) before a task times out and fails.
	TimeoutMs int

	// Amount of attempts a worker has tried to process the task
	Attempts int

	// Time in ms that the task was claimed
	ClaimToken string
}
