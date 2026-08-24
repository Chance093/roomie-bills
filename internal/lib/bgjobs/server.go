package bgjobs

import (
	"context"
	"encoding/json"
	"fmt"
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

type Handler func(Task) error

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
	ctx := context.Background()

	// run multiple workers concurrently
	var wg sync.WaitGroup
	for i := 1; i <= s.cfg.Concurrency; i++ {
		wg.Go(func() {
			worker(ctx, mux, rdb, i)
		})
	}

	wg.Wait()
}

func worker(ctx context.Context, mux *ServeMux, rdb *redis.Client, i int) error {
	for {
		fmt.Printf("Worker %d listening for tasks...\n", i)
		// dequeue task (blocking)
		// TODO: Reliable queue (send to temp queue until confirmed done)
		// TODO: change this to BLMove
		raw, err := rdb.BRPop(ctx, time.Duration(0), Queue).Result()
		if err != nil {
			return fmt.Errorf("Failed to dequeue: %w", err)
		}

		// unmarshall json
		var task Task
		if err := json.Unmarshal([]byte(raw[1]), &task); err != nil {
			return fmt.Errorf("Failed to unmarshal raw task json: %w", err)
		}

		// look up task name in mux
		h, err := mux.getHandler(task.Name)
		if err != nil {
			return fmt.Errorf("Failed to get task handler: %w", err)
		}

		// TODO: Implement task timeout (force kill and retry)
		// TODO: Implement backoff retry logic
		// TODO: DLQ (when task runs out of retries, send to DLQ)
		if err := h(task); err != nil {
			return fmt.Errorf("Task handler failed: %w", err)
		}
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
