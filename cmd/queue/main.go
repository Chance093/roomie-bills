package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
)

const (
	Addr         = "127.0.0.1:6379"
	TypeTestTask = "test:task"
)

type TaskPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// processor
	client := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})

	payload := TaskPayload{
		Name: "Chance",
		Age:  27,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	// push tasks to queue
	task1 := bgjobs.NewTask(TypeTestTask, b, bgjobs.MaxRetry(5))
	if _, err := client.Enqueue(task1); err != nil {
		log.Fatal(err)
	}

	// worker
	srv := bgjobs.NewServer(bgjobs.RedisOpts{Addr: Addr}, bgjobs.ServerConfig{Concurrency: 5})
	mux := bgjobs.NewServeMux()
	mux.HandleFunc(TypeTestTask, sendNameAndAge)

	srv.Run(mux)
}

func sendNameAndAge(ctx context.Context, t bgjobs.Task) error {
	return fmt.Errorf("Error in task: %s", t.Name)
}
