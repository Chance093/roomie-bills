package taskqueue

import "context"

type Server struct {
	pool      *WorkerPool
	parentCtx context.Context
}

type ServerOpts struct {
	Concurrency int
	TaskQueueOpts
	WorkerOpts
}

func NewServer(ctx context.Context, opts ServerOpts) (Server, error) {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}

	queue := newTaskQueue(opts.TaskQueueOpts)

	pool, err := NewWorkerPool(ctx, opts.Concurrency, queue, opts.WorkerOpts)
	if err != nil {
		return Server{}, err
	}

	return Server{pool: pool, parentCtx: ctx}, nil
}

func (s *Server) Run(mux ServeMux) {
	s.pool.Start(s.parentCtx, mux)
}
