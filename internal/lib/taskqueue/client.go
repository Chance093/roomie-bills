package taskqueue

import "context"

type Client struct {
	queue     *taskQueue
	parentCtx context.Context
}

type ClientOpts struct {
	TaskQueueOpts
}

func NewClient(ctx context.Context, opts ClientOpts) Client {
	queue := newTaskQueue(opts.TaskQueueOpts)

	return Client{parentCtx: ctx, queue: queue}
}

func (c Client) Enqueue(task Task) (string, error) {
	taskMeta := NewTaskMeta(task)
	return c.queue.Enqueue(c.parentCtx, taskMeta)
}
