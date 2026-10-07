package taskqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type worker struct {
	name  string
	queue *taskQueue
	mux   *ServeMux

	workerOpts

	// amount of tasks processed by worker
	processedN atomic.Int64

	runMu  sync.Mutex
	done   chan struct{}
	cancel context.CancelFunc
}

type workerOpts struct {
	// amount of time a worker will wait to claim a task from queue.
	// default: 1000
	claimTimeoutMs int
}

// Creates a new worker which will pull tasks off the task queue and process them.
// To start the worker, try running worker.start().
func newWorker(name string, queue *taskQueue, opts workerOpts) (*worker, error) {
	if queue == nil {
		return nil, errors.New("Queue must be provided to worker")
	}

	w := &worker{
		name:       name,
		queue:      queue,
		workerOpts: opts,
	}

	w.setDefaultOpts()

	return w, nil
}

// sets default options for worker
func (w *worker) setDefaultOpts() {
	if w.claimTimeoutMs < 1000 {
		w.claimTimeoutMs = 1000
	}
}

// Will start a worker by creating channels to communicate when to stop,
// and running the worker by using worker.run(). Works concurrently.
func (w *worker) start(ctx context.Context, mux *ServeMux) {
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

	w.mux = mux

	go w.run(ctx, done)
}

// Will stop a worker by calling cancel context. Works concurrently.
func (w *worker) stop() {
	w.runMu.Lock()
	cancel := w.cancel
	w.runMu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// Checks if worker is alive by checking the done channel. If done channel
// does not have a value, then the worker is alive. Works concurrently.
func (w *worker) isAlive() bool {
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

// Returns the amount of tasks processed by the worker.
func (w *worker) processed() int64 {
	return w.processedN.Load()
}

// Resets the amount of tasks that were processed by the worker.
func (w *worker) resetProcessed() {
	w.processedN.Store(0)
}

// Will loop forever, checking if a task can be claimed by the queue. If
// task was claimed, then we run worker.process(). If no task is given to
// us, then we continue the loop. We accept nil tasks so that we can
// continuously check if a parent context has been cancelled.
func (w *worker) run(ctx context.Context, done chan struct{}) {
	defer close(done) // when done running, close done channel for IsAlive method

	for {
		select {
		case <-ctx.Done(): // Stop() was called, so return
			return
		default: // do nothing
		}

		task, err := w.queue.claim(ctx, w.claimTimeoutMs)
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

// Will process the claimed task by getting the task handler, setting a
// timeout and some channels for goroutine communication, and blocking
// until we receive our result on a channel. Then we move task to either
// completed or failed.
func (w *worker) process(parentCtx context.Context, task *ClaimedTask) {
	// look up task handler in mux
	h, err := w.mux.getHandler(task.Name)
	if err != nil {
		if _, err := w.queue.fail(parentCtx, task, err); err != nil {
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
		if _, err := w.queue.fail(parentCtx, task, err); err != nil {
			fmt.Printf("Error while trying to fail task: %s", err.Error())
		}
		return
	}

	// complete task
	if ok, err := w.queue.complete(parentCtx, task); err == nil && ok {
		w.processedN.Add(1)
	} else {
		fmt.Printf("Error while trying to complete task: %s", err.Error())
	}
}
