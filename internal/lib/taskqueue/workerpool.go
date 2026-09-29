package taskqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type WorkerPool struct {
	workers []*worker

	WorkerOpts

	queue     *taskQueue
	mux       ServeMux
	parentCtx context.Context

	mu sync.Mutex
}

func NewWorkerPool(ctx context.Context, count int, queue *taskQueue, opts WorkerOpts) (*WorkerPool, error) {
	if queue == nil {
		return nil, errors.New("Queue must be provided to worker pool")
	}

	pool := &WorkerPool{
		queue:      queue,
		parentCtx:  ctx,
		WorkerOpts: opts,
	}

	if err := pool.Resize(count); err != nil {
		return nil, err
	}

	return pool, nil
}

func (p *WorkerPool) Resize(size int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.workers) < size {
		workerName := fmt.Sprintf("Worker - %d", len(p.workers)+1)
		worker, err := NewWorker(workerName, p.queue, p.mux, p.WorkerOpts)
		if err != nil {
			return err
		}

		p.workers = append(p.workers, worker)
	}
	for len(p.workers) > size {
		workerIdx := len(p.workers) - 1
		p.workers[workerIdx].Stop()
		p.workers = p.workers[:workerIdx]
	}

	return nil
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
