package taskqueue

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// server handles the initialization and orchestration of workers, queues,
// and the processing of tasks. It is meant to be used as a consumer (as
// opposed to a producer or consumer). Initialized and run in a
// process you intend to handle asynchronous tasks.
type server struct {
	pool      *workerPool
	queue     *taskQueue
	parentCtx context.Context
}

type ServerOpts struct {
	// amount of concurrent workers.
	// default: 1
	Concurrency int
	// amount of time a worker will wait to claim a task from queue.
	// default: 1000
	ClaimTimeoutMs int
}

// Creates a new server which initializes a task queue along with workers.
// Use server.Run() to start the workers, which will pull tasks off the task
// queue to start processing the work. If you would like more fine control of
// how the queue works, try creating a custom queue with NewTaskQueue(), and
// passing the returned queue to NewServerWithCustomQueue().
func NewServer(ctx context.Context, opts ServerOpts) server {
	queue := NewTaskQueue(TaskQueueOpts{}) // create default queue

	return newServer(ctx, queue, opts)
}

// Creates a new server with the custom queue that was passed in. To create
// a custom queue, try using NewTaskQueue(). If no queue is passed in, will
// initialize default queue. It initializes workers in the same way as
// NewServer(). Use server.Run() to start the workers, which will pull tasks
// off the task queue to start processing the work.
func NewServerWithCustomQueue(ctx context.Context, queue *taskQueue, opts ServerOpts) server {
	if queue == nil {
		queue = NewTaskQueue(TaskQueueOpts{}) // create default queue
	}

	return newServer(ctx, queue, opts)
}

func newServer(ctx context.Context, queue *taskQueue, opts ServerOpts) server {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}

	pool := newWorkerPool(ctx, opts.Concurrency, queue, workerOpts{opts.ClaimTimeoutMs})

	return server{pool: pool, parentCtx: ctx}
}

// Starts the worker pool, which will allow workers to start pulling off tasks
// from the task queue to process the work. Will run until the process is
// interrupted.
func (s *server) Run(mux *ServeMux) {
	s.pool.start(s.parentCtx, mux)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Print("shutting down")
	s.pool.stop()
}
