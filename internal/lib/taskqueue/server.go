package taskqueue

import "context"

type Server struct {
	pool      *WorkerPool
	parentCtx context.Context
}

type ServerOpts struct {
	Concurrency int
	TaskQueueOptions
}

func NewServer(ctx context.Context, opts *ServerOpts) Server {
	if opts == nil {
		opts = &ServerOpts{}
	}

	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}

	queue := newTaskQueue(opts.TaskQueueOptions)

	pool := NewWorkerPool(ctx, opts.Concurrency, queue, nil)

	return Server{pool: pool, parentCtx: ctx}
}

func (s *Server) Run(mux ServeMux) {
	s.pool.Start(s.parentCtx, mux)
}
