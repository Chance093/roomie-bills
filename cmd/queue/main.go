package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
)

const (
	Addr         = "127.0.0.1:6379"
	TypeTestTask = "test:task"
)

func main() {
	// processor
	client := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})

	// push tasks to queue
	task1 := bgjobs.NewTask(TypeTestTask, []byte("test1"))
	task2 := bgjobs.NewTask(TypeTestTask, []byte("test2"))
	task3 := bgjobs.NewTask(TypeTestTask, []byte("test3"))
	task4 := bgjobs.NewTask(TypeTestTask, []byte("test4"))
	task5 := bgjobs.NewTask(TypeTestTask, []byte("test5"))
	if _, err := client.Enqueue(task1); err != nil {
		log.Fatal(err)
	}
	if _, err := client.Enqueue(task2); err != nil {
		log.Fatal(err)
	}
	if _, err := client.Enqueue(task3); err != nil {
		log.Fatal(err)
	}
	if _, err := client.Enqueue(task4); err != nil {
		log.Fatal(err)
	}
	if _, err := client.Enqueue(task5); err != nil {
		log.Fatal(err)
	}

	// worker
	srv := bgjobs.NewServer(bgjobs.RedisOpts{Addr: Addr}, bgjobs.ServerConfig{Concurrency: 5})
	mux := bgjobs.NewServeMux()
	mux.HandleFunc(TypeTestTask, testHandler)

	srv.Run(mux)
}

func testHandler(t bgjobs.Task) error {
	fmt.Println("Working task...")
	var buffer bytes.Buffer
	buffer.Write(t.Payload)
	time.Sleep(time.Second * 3)
	fmt.Println(buffer.String())
	fmt.Println("Finished task...")
	return nil
}
