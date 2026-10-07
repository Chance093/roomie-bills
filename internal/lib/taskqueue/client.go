package taskqueue

import "context"

// Client handles the enqueueing of tasks, which will later be processed
// by workers. Client is meant to be used as a producer of tasks (as
// opposed to a broker or consumer).
type Client struct {
	queue     *taskQueue
	parentCtx context.Context
}

// Creates a new client which initializes a task queue. Use client.Enqueue()
// to enqueue a task, which will be pulled off the queue by workers. If you
// would like more fine control of how the queue works, try creating a
// custom queue with NewTaskQueue(), and passing the returned queue to
// NewClientWithCustomQueue().
func NewClient(ctx context.Context) Client {
	queue := NewTaskQueue(TaskQueueOpts{}) // create default queue

	return Client{parentCtx: ctx, queue: queue}
}

// Creates a new client with the custom queue that was passed in. To create
// a custom queue, try using NewTaskQueue(). If no queue is passed in, will
// initialize default queue. Use client.Enqueue() to enqueue a task, which
// will be pulled off the queue by workers
func NewClientWithCustomQueue(ctx context.Context, queue *taskQueue) Client {
	if queue == nil {
		queue = NewTaskQueue(TaskQueueOpts{}) // create default queue
	}

	return Client{parentCtx: ctx, queue: queue}
}

// Enqueues a task to be processed by a worker. To create a task, try
// NewTask().
func (c Client) Enqueue(task Task) (string, error) {
	taskMeta := newTaskMeta(task)
	return c.queue.enqueue(c.parentCtx, taskMeta)
}
