package bgjobs

type Task struct {
	Name    string `json:"name"`
	Payload []byte `json:"payload"`
	Retries int    `json:"retries"`
	Timeout int    `json:"timeout"`
}

func NewTask(name string, payload []byte) *Task {
	return &Task{
		Name:    name,
		Payload: payload,
		Retries: 5,
		Timeout: 20,
	}
}
