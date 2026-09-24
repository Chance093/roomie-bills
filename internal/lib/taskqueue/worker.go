package taskqueue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type worker struct {
	name  string
	queue *taskQueue
	mux   ServeMux

	workerOpts

	processed atomic.Int64

	runMu  sync.Mutex
	done   chan struct{}
	cancel context.CancelFunc
}

type workerOpts struct {
	claimTimeoutMs int
}

func NewWorker(name string, queue *taskQueue, mux ServeMux, opts *workerOpts) *worker {
	if opts == nil {
		opts = &workerOpts{}
	}

	if opts.claimTimeoutMs <= 0 {
		opts.claimTimeoutMs = 500
	}

	return &worker{name: name, queue: queue, mux: mux, workerOpts: *opts}
}

// Create done channel and cancel context
// If its already running, nop
func (w *worker) Start(ctx context.Context) {
	w.runMu.Lock()
	defer w.runMu.Unlock()

	if w.done != nil {
		select {
		case <-w.done:
			// continue on to rerun worker
		default:
			return // else don't do anything (worker already running)
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	done := make(chan struct{})
	w.done = done

	go w.run(ctx, done)
}

func (w *worker) Stop() {
	w.runMu.Lock()
	cancel := w.cancel
	w.runMu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (w *worker) IsAlive() bool {
	w.runMu.Lock()
	done := w.done
	w.runMu.Unlock()

	if done == nil {
		return false
	}

	select {
	case <-done:
		return false
	default:
		return true
	}
}

func (w *worker) Processed() int64 {
	return w.processed.Load()
}

func (w *worker) ResetProcessed() {
	w.processed.Store(0)
}

// constantly tries to pull task off of queue
func (w *worker) run(ctx context.Context, done chan struct{}) {
	defer close(done) // when done running, close done channel for IsAlive method

	for {
		select {
		case <-ctx.Done(): // Stop() was called, so return
			return
		default: // do nothing
		}

		task, err := w.queue.Claim(ctx, w.claimTimeoutMs)
		if err != nil {
			select {
			case <-ctx.Done(): // Stop() was called, so return
				return
			default:
				fmt.Println(err.Error())           // print error
				time.Sleep(time.Millisecond * 100) // back off
				continue
			}
		}

		if task == nil {
			continue // no need for sleep, Claim() handles timeout
		}

		w.process(ctx, task)
	}
}

// processes job
func (w *worker) process(parentCtx context.Context, task *ClaimedTask) {
	// look up task handler in mux
	h, err := w.mux.getHandler(task.Name)
	if err != nil {
		if _, err := w.queue.Fail(parentCtx, task, err); err != nil {
			fmt.Printf("Error while trying to fail task: %s", err.Error())
		}
		return
	}

	// execute task handler
	// TODO: (REVIEW) Is this the right way to use timeout or should I just pass ctx with no channels
	timeout := time.Duration(task.TimeoutMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()
	errChan := make(chan error)
	doneChan := make(chan struct{})

	go func() {
		h(ctx, task, errChan, doneChan)
	}()

	// block until we get result of handler
	select {
	case <-ctx.Done(): // timed out
		err = fmt.Errorf("Timed out while running task handler for task (%s): %w", task.Id, ctx.Err())
	case e := <-errChan: // error
		err = e
	case <-doneChan: // success
	}

	// if error, fail task
	if err != nil {
		if _, err := w.queue.Fail(parentCtx, task, err); err != nil {
			fmt.Printf("Error while trying to fail task: %s", err.Error())
		}
		return
	}

	// complete task
	if ok, err := w.queue.Complete(parentCtx, task); err == nil && ok {
		w.processed.Add(1)
	} else {
		fmt.Printf("Error while trying to complete task: %s", err.Error())
	}
}
