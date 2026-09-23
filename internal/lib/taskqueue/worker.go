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

	processN atomic.Int32

	runMu  sync.Mutex
	done   chan struct{}
	cancel context.CancelFunc
}

func NewWorker(name string, queue *taskQueue) *worker {
	return &worker{name: name, queue: queue}
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

	go w.Run(ctx, done)
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

// constantly tries to pull task off of queue
func (w *worker) Run(ctx context.Context, done chan struct{}) {
	defer close(done) // when done running, close done channel for IsAlive method

	for {
		select {
		case <-ctx.Done(): // Stop() was called, so return
			return
		default: // do nothing
		}

		task, err := w.queue.Claim(ctx, 500)
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

		w.Process(ctx, task)
	}
}

// processes job
func (w *worker) Process(ctx context.Context, task *ClaimedTask) {
	// look up task handler in mux
}
