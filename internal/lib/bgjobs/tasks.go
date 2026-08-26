package bgjobs

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Id      string        `json:"id"`
	Name    string        `json:"name"`
	Payload []byte        `json:"payload"`
	Retries int8          `json:"retries"`
	Timeout time.Duration `json:"timeout"`
}

type Timeout int64

type MaxRetry int8

func NewTask(name string, payload []byte, opts ...any) *Task {
	t := &Task{
		Id:      uuid.New().String(),
		Name:    name,
		Payload: payload,
		Retries: 1,
		Timeout: time.Second * 5,
	}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case Timeout:
			timeout := time.Duration(opt)
			if timeout > time.Duration(0) {
				t.Timeout = timeout
			}
			continue
		case MaxRetry:
			retry := int8(opt)
			if retry > 0 {
				t.Retries = retry
			}
			continue
		default:
			continue
		}
	}

	return t
}
