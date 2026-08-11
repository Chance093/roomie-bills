package bgjobs

import "github.com/google/uuid"

type Task struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Payload []byte `json:"payload"`
	Retries int    `json:"retries"`
	Timeout int    `json:"timeout"`
}

func NewTask(name string, payload []byte) *Task {
	return &Task{
		Id:      uuid.New().String(),
		Name:    name,
		Payload: payload,
		Retries: 5,
		Timeout: 20,
	}
}
