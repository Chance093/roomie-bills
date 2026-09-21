package taskqueue

import (
	"github.com/google/uuid"
)

type Task struct {
	Id            string `redis:"id"`
	Name          string `redis:"name"`
	Payload       string `redis:"payload"`
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

func NewTask(name string, payload []byte) *Task {
	t := &Task{
		Id:      uuid.New().String(),
		Name:    name,
		Payload: string(payload),
	}

	return t
}
