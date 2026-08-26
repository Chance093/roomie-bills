package bgjobs

import (
	"testing"
	"time"
)

type TestNewTaskInput struct {
	name    string
	payload []byte
	opts    []any
}

func TestNewTask(t *testing.T) {
	tests := []struct {
		name  string
		input TestNewTaskInput
		want  Task
	}{
		{
			name: "Normal Task",
			input: TestNewTaskInput{
				name:    "Task",
				payload: []byte("Task 1"),
			},
			want: Task{
				Name:    "Task",
				Payload: []byte("Task 1"),
				Retries: 1,
				Timeout: time.Second * 5,
			},
		},
		{
			name: "Task with custom retries",
			input: TestNewTaskInput{
				name:    "Task",
				payload: []byte("Hello"),
				opts:    []any{MaxRetry(5)},
			},
			want: Task{
				Name:    "Task",
				Payload: []byte("Hello"),
				Retries: 5,
				Timeout: time.Second * 5,
			},
		},
		{
			name: "Task with custom timeout",
			input: TestNewTaskInput{
				name:    "Task",
				payload: []byte("Hello"),
				opts:    []any{Timeout(time.Second * 3)},
			},
			want: Task{
				Name:    "Task",
				Payload: []byte("Hello"),
				Retries: 1,
				Timeout: time.Second * 3,
			},
		},
		{
			name: "Task with custom timeout and retries",
			input: TestNewTaskInput{
				name:    "Task",
				payload: []byte("Hello"),
				opts:    []any{Timeout(time.Second * 6), MaxRetry(5)},
			},
			want: Task{
				Name:    "Task",
				Payload: []byte("Hello"),
				Retries: 5,
				Timeout: time.Second * 6,
			},
		},
		{
			name: "Task with bad timeout and retries",
			input: TestNewTaskInput{
				name:    "Task",
				payload: []byte("Hello"),
				opts:    []any{Timeout(time.Duration(0)), MaxRetry(0)},
			},
			want: Task{
				Name:    "Task",
				Payload: []byte("Hello"),
				Retries: 1,
				Timeout: time.Second * 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask(tt.input.name, tt.input.payload, tt.input.opts...)

			if task.Name != tt.want.Name {
				t.Errorf("Task name not as expected. want: %s, got: %s", tt.want.Name, task.Name)
			}
			if string(task.Payload) != string(tt.want.Payload) {
				t.Errorf("Task payload not as expected. want: %s, got: %s", string(tt.want.Payload), string(task.Payload))
			}
			if task.Retries != tt.want.Retries {
				t.Errorf("Task retries not as expected. want: %d, got: %d", tt.want.Retries, task.Retries)
			}
			if task.Timeout != tt.want.Timeout {
				t.Errorf("Task timeout not as expected. want: %d, got: %d", tt.want.Timeout, task.Timeout)
			}
		})
	}
}

func TestSetTaskOpts(t *testing.T) {
	task := Task{
		Retries: 1,
		Timeout: time.Second * 5,
	}

	tests := []struct {
		name string
		opts []any
		want Task
	}{
		{
			name: "No opts",
			want: Task{
				Retries: 1,
				Timeout: time.Second * 5,
			},
		},
		{
			name: "Set retries",
			opts: []any{MaxRetry(5)},
			want: Task{
				Retries: 5,
				Timeout: time.Second * 5,
			},
		},
		{
			name: "Set timeout",
			opts: []any{Timeout(time.Second * 2)},
			want: Task{
				Retries: 1,
				Timeout: time.Second * 2,
			},
		},
		{
			name: "Set retries and timeout",
			opts: []any{MaxRetry(5), Timeout(time.Second * 2)},
			want: Task{
				Retries: 5,
				Timeout: time.Second * 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copy := task
			setTaskOpts(&copy, tt.opts)

			if copy.Retries != tt.want.Retries {
				t.Errorf("Task retries not as expected. want: %d, got: %d", tt.want.Retries, copy.Retries)
			}
			if copy.Timeout != tt.want.Timeout {
				t.Errorf("Task timeout not as expected. want: %d, got: %d", tt.want.Timeout, copy.Timeout)
			}
		})
	}
}
