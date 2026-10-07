package taskqueue

import (
	"context"
	"fmt"
)

type (
	Handler             func(context.Context, *claimedTask) error
	HandlerWithChannels func(context.Context, *claimedTask, chan error, chan struct{})
)

type ServeMux struct {
	m map[string]HandlerWithChannels
}

// Creates a new Server Multiplexer for mapping task names to task handlers.
// To map a task name to a task handler, try running ServeMux.HandleFunc().
func NewServeMux() *ServeMux {
	return &ServeMux{
		m: make(map[string]HandlerWithChannels),
	}
}

// Maps a task name to a task handler. When a task gets pulled off of the
// task queue, the task name will tell it which handler it needs to run.
func (m *ServeMux) HandleFunc(taskName string, handler Handler) {
	m.m[taskName] = handlerToHandlerWithChannels(handler)
}

// Grabs the task handler that is mapped to the taskName that is passed in.
func (m *ServeMux) getHandler(taskName string) (HandlerWithChannels, error) {
	h, ok := m.m[taskName]
	if !ok {
		return nil, fmt.Errorf("Handler does not exist for task type: %s", taskName)
	}

	return h, nil
}

// Higher order helper which takes a Handler, and returns another handler
// that uses communication channels. This allows all handlers to be run as
// concurrent goroutines, without the implementer of a handler having to
// deal with the channel communication.
func handlerToHandlerWithChannels(handler Handler) HandlerWithChannels {
	return func(ctx context.Context, t *claimedTask, errChan chan error, doneChan chan struct{}) {
		if err := handler(ctx, t); err != nil {
			errChan <- err
		} else {
			doneChan <- struct{}{}
		}
	}
}
