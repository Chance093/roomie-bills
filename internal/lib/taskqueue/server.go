package taskqueue

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	pool      *WorkerPool
	parentCtx context.Context
}

type ServerOpts struct {
	Concurrency int
	TaskQueueOpts
	WorkerOpts
}

func NewServer(ctx context.Context, opts ServerOpts) Server {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}

	queue := newTaskQueue(opts.TaskQueueOpts)

	pool, _ := NewWorkerPool(ctx, opts.Concurrency, queue, opts.WorkerOpts)

	return Server{pool: pool, parentCtx: ctx}
}

func (s *Server) Run(mux *ServeMux) {
	s.pool.Start(s.parentCtx, mux)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Print("shutting down")
	s.pool.Stop()
}
