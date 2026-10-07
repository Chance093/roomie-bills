package taskqueue

import (
	"context"
	"fmt"
	"sync"
)

type workerPool struct {
	workers []*worker

	workerOpts

	queue     *taskQueue
	mux       *ServeMux
	parentCtx context.Context

	mu sync.Mutex
}

func newWorkerPool(ctx context.Context, count int, queue *taskQueue, opts workerOpts) *workerPool {
	pool := &workerPool{
		queue:      queue,
		parentCtx:  ctx,
		workerOpts: opts,
	}

	if err := pool.resize(count); err != nil {
		return nil
	}

	return pool
}

func (p *workerPool) resize(size int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.workers) < size {
		workerName := fmt.Sprintf("Worker - %d", len(p.workers)+1)
		worker, err := newWorker(workerName, p.queue, p.workerOpts)
		if err != nil {
			return err
		}

		p.workers = append(p.workers, worker)
	}
	for len(p.workers) > size {
		workerIdx := len(p.workers) - 1
		p.workers[workerIdx].stop()
		p.workers = p.workers[:workerIdx]
	}

	return nil
}

func (p *workerPool) start(ctx context.Context, mux *ServeMux) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.mux = mux
	for _, worker := range p.workers {
		worker.start(ctx, p.mux)
	}
}

func (p *workerPool) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, worker := range p.workers {
		worker.stop()
	}
}

func (p *workerPool) running() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	var n int
	for _, worker := range p.workers {
		if worker.isAlive() {
			n++
		}
	}

	return n
}

func (p *workerPool) processed() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	var n int64
	for _, worker := range p.workers {
		n += worker.processed()
	}

	return n
}

func (p *workerPool) resetProcessed() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, worker := range p.workers {
		worker.resetProcessed()
	}
}
