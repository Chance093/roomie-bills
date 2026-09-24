package taskqueue

import (
	"context"
	"fmt"
	"sync"
)

type WorkerPool struct {
	workers []*worker

	workerPoolOpts

	queue     *taskQueue
	mux       ServeMux
	parentCtx context.Context

	mu sync.Mutex
}

type workerPoolOpts struct {
	claimTimeoutMs int
}

func NewWorkerPool(ctx context.Context, count int, queue *taskQueue, opts *workerPoolOpts) *WorkerPool {
	if opts == nil {
		opts = &workerPoolOpts{}
	}

	pool := &WorkerPool{
		queue:          queue,
		parentCtx:      ctx,
		workerPoolOpts: *opts,
	}

	pool.Resize(count)

	return pool
}

func (p *WorkerPool) Resize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.workers) < size {
		workerName := fmt.Sprintf("Worker - %d", len(p.workers)+1)
		worker := NewWorker(workerName, p.queue, p.mux, (*workerOpts)(&p.workerPoolOpts))
		p.workers = append(p.workers, worker)
	}
	for len(p.workers) > size {
		workerIdx := len(p.workers) - 1
		p.workers[workerIdx].Stop()
		p.workers = p.workers[:workerIdx]
	}
}

func (p *WorkerPool) Start(ctx context.Context, mux ServeMux) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.mux = mux
	for _, worker := range p.workers {
		worker.Start(ctx)
	}
}

func (p *WorkerPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, worker := range p.workers {
		worker.Stop()
	}
}

func (p *WorkerPool) Running() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	var n int
	for _, worker := range p.workers {
		if worker.IsAlive() {
			n++
		}
	}

	return n
}

func (p *WorkerPool) Processed() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	var n int64
	for _, worker := range p.workers {
		n += worker.Processed()
	}

	return n
}

func (p *WorkerPool) ResetProcessed() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, worker := range p.workers {
		worker.ResetProcessed()
	}
}
