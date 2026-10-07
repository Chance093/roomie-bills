package taskqueue

import (
	"github.com/google/uuid"
)

type TaskMeta struct {
	Id            string `redis:"id"`
	Name          string `redis:"name"`
	Payload       string `redis:"payload"`
	TimeoutMs     int    `redis:"timeout_ms"`
	Status        string `redis:"status"`
	Attempts      int    `redis:"attempts"`
	EnqueuedAtMs  int64  `redis:"enqueued_at_ms"`
	ClaimedAtMs   int64  `redis:"claimed_at_ms"`
	CompletedAtMs int64  `redis:"completed_at_ms"`
	ReclaimedAtMs int64  `redis:"reclaimed_at_ms"`
	ClaimToken    string `redis:"claim_token"`
	LastError     string `redis:"last_error"`
	LastErrorAtMs int64  `redis:"last_error_at_ms"`
}

type ClaimedTask struct {
	Id         string
	Name       string
	Payload    string
	TimeoutMs  int
	Attempts   int
	ClaimToken string
}

func NewTaskMeta(task Task) TaskMeta {
	return TaskMeta{
		Id:        uuid.New().String(),
		Name:      task.Name,
		Payload:   task.Payload,
		TimeoutMs: task.TimeoutMs,
	}
}

type Task struct {
	Name      string
	Payload   string
	TimeoutMs int
}

type TaskOpts struct {
	TimeoutMs int
}

func NewTask(name string, payload []byte, opts TaskOpts) Task {
	t := Task{
		Name:      name,
		Payload:   string(payload),
		TimeoutMs: opts.TimeoutMs,
	}

	t.setDefaultOpts()

	return t
}

func (t *Task) setDefaultOpts() {
	if t.TimeoutMs <= 0 {
		t.TimeoutMs = 5000 // 5 seconds
	}
}
