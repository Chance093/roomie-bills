package bgjobs

import (
	"context"
	"fmt"
)

type Handler func(context.Context, Task) error

// multiplexer that maps task names to task handlers
type ServeMux struct {
	m map[string]Handler
}

// return ServeMux struct
func NewServeMux() *ServeMux {
	return &ServeMux{
		m: make(map[string]Handler),
	}
}

// maps task name to a task handler
func (m *ServeMux) HandleFunc(taskType string, handler Handler) {
	m.m[taskType] = handler
}

func (m *ServeMux) getHandler(taskType string) (Handler, error) {
	h, ok := m.m[taskType]
	if !ok {
		return nil, fmt.Errorf("Handler does not exist for task type: %s", taskType)
	}

	return h, nil
}
