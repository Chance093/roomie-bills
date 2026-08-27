package bgjobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"
)

type Server struct {
	cfg    ServerConfig
	primeQ queue
	tempQ  queue
	dlq    queue
}

type ServerConfig struct {
	Concurrency int
	MaxTimeout  time.Duration
}

// set options to server
func NewServer(opts RedisOpts, cfg ServerConfig) Server {
	rdb := initNewRDBClient(opts)
	pq := newPrimaryQueue(rdb)
	tq := newTempQueue(rdb)
	dlq := newDLQ(rdb)

	// set default cfg
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 1
	}
	if cfg.MaxTimeout == time.Duration(0) {
		cfg.MaxTimeout = time.Duration(10 * time.Second)
	}

	return Server{cfg, pq, tq, dlq}
}

// continuously tries to pop off queue until task shows up,
// then it will look up the task name in the multiplexer and
// get the handler to run
func (s *Server) Run(mux *ServeMux) {
	defer s.primeQ.rdb.Close()
	defer s.tempQ.rdb.Close()
	defer s.dlq.rdb.Close()

	// run multiple workers concurrently
	var wg sync.WaitGroup
	for i := 1; i <= s.cfg.Concurrency; i++ {
		wg.Go(func() {
			s.worker(mux)
		})
	}

	wg.Wait()
}

// TODO: error logging
func (s *Server) worker(mux *ServeMux) error {
OUTER:
	for {
		ctx := context.TODO()

		// dequeue task (blocking) and move to temp queue (reliable queue)
		raw, err := s.primeQ.popAndMoveTo(ctx, s.tempQ)
		if err != nil {
			return fmt.Errorf("Failed to dequeue and move to temp queue: %w", err)
		}

		// unmarshall json into Task struct
		var t Task
		if err := json.Unmarshal([]byte(raw), &t); err != nil {
			return fmt.Errorf("Failed to unmarshal raw task json: %w", err)
		}

		// look up task name in mux
		h, err := mux.getHandler(t.Name)
		if err != nil {
			return fmt.Errorf("Failed to get task handler: %w", err)
		}

		// run handler with backoff retries
		var errMessages []string
		for attempt := 1; attempt <= int(t.Retries); attempt++ {
			// TODO: Implement timeout logic
			ctx, cancel := context.WithTimeout(ctx, t.Timeout)
			defer cancel()

			err := h(ctx, t)
			if err == nil {
				// on success, remove task from temp queue
				if err := s.tempQ.remove(ctx, raw); err != nil {
					return fmt.Errorf("Failed to remove task from temp queue: %w\n", err)
				}
				continue OUTER
			}

			// backoff before retry
			errMessages = append(errMessages, err.Error())
			if attempt < int(t.Retries) {
				sleepDur := time.Duration(math.Pow(2, float64(attempt)))
				time.Sleep(time.Second * sleepDur)
			}
		}

		// add to DLQ if retries run out and remove from temp queue
		if err := sendToDLQ(ctx, s.dlq, t, errMessages); err != nil {
			newErr := fmt.Errorf("Failed to send task to DLQ: %w\n", err)
			fmt.Println(newErr)
			return newErr
		}

		if err := s.tempQ.remove(ctx, raw); err != nil {
			return fmt.Errorf("Failed to remove task from temp queue: %w\n", err)
		}
	}
}
