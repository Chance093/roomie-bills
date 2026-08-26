package bgjobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type ServerConfig struct {
	Concurrency int
	MaxTimeout  string
}

type Server struct {
	RedisOpts
	cfg ServerConfig
}

type Handler func(context.Context, Task) error

// set options to server
func NewServer(opts RedisOpts, cfg ServerConfig) Server {
	return Server{opts, cfg}
}

// continuously tries to pop off queue until task shows up,
// then it will look up the task name in the multiplexer and
// get the handler to run
func (s *Server) Run(mux *ServeMux) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     s.Addr,
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})

	// run multiple workers concurrently
	var wg sync.WaitGroup
	for i := 1; i <= s.cfg.Concurrency; i++ {
		wg.Go(func() {
			worker(mux, rdb)
		})
	}

	wg.Wait()
}

func worker(mux *ServeMux, rdb *redis.Client) error {
OUTER:
	for {
		ctx := context.Background()

		// dequeue task (blocking) and move to temp queue (reliable queue)
		raw, err := rdb.BLMove(ctx, Queue, Temp, "RIGHT", "LEFT", time.Duration(0)).Result()
		if err != nil {
			return fmt.Errorf("Failed to dequeue: %w", err)
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

		// TODO: Implement timeout logic
		ctx, cancel := context.WithTimeout(ctx, t.Timeout)
		defer cancel()

		// run handler with backoff retries
		for attempt := 1; attempt <= int(t.Retries); attempt++ {
			err := h(ctx, t)
			if err == nil {
				// on success, remove task from temp queue
				if _, err := rdb.LRem(ctx, Temp, 0, raw).Result(); err != nil {
					return fmt.Errorf("Failed to remove task from temp queue: %w\n", err)
				}
				continue OUTER
			}

			// back off before retry
			fmt.Println(err)
			sleepDur := time.Duration(math.Pow(2, float64(attempt)))
			time.Sleep(time.Second * sleepDur)
		}

		// code only executes if retries ran out
		// TODO: Implement DLQ (when task runs out of retries, send to DLQ)
	}
}

// multiplexer that maps task names to task handlers
type ServeMux struct {
	m map[string]Handler
}

// return ServeMux struct
func NewServeMux() *ServeMux {
	return &ServeMux{
		m: make(map[string]Handler),
	}
}

// maps task name to a task handler
func (m *ServeMux) HandleFunc(taskType string, handler Handler) {
	m.m[taskType] = handler
}

func (m *ServeMux) getHandler(taskType string) (Handler, error) {
	h, ok := m.m[taskType]
	if !ok {
		return nil, fmt.Errorf("Handler does not exist for task type: %s", taskType)
	}

	return h, nil
}
