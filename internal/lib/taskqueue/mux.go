package taskqueue

import (
	"context"
	"fmt"
)

type (
	Handler             func(context.Context, *claimedTask) error
	HandlerWithChannels func(context.Context, *claimedTask, chan error, chan struct{})
)

// multiplexer that maps task names to task handlers
type ServeMux struct {
	m map[string]HandlerWithChannels
}

// return ServeMux struct
func NewServeMux() *ServeMux {
	return &ServeMux{
		m: make(map[string]HandlerWithChannels),
	}
}

// maps task name to a task handler
func (m *ServeMux) HandleFunc(taskType string, handler Handler) {
	m.m[taskType] = handlerToHandlerWithChannels(handler)
}

func (m *ServeMux) getHandler(taskType string) (HandlerWithChannels, error) {
	h, ok := m.m[taskType]
	if !ok {
		return nil, fmt.Errorf("Handler does not exist for task type: %s", taskType)
	}

	return h, nil
}

func handlerToHandlerWithChannels(handler Handler) HandlerWithChannels {
	return func(ctx context.Context, t *claimedTask, errChan chan error, doneChan chan struct{}) {
		if err := handler(ctx, t); err != nil {
			errChan <- err
		} else {
			doneChan <- struct{}{}
		}
	}
}
