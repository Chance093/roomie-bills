package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
)

const (
	Addr           = "127.0.0.1:6379"
	TypeBankDelete = "delete:bank"
	TypeBankAdd    = "add:bank"
	TypeBankUpdate = "update:bank"
)

func main() {
	// processor
	client := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})

	// push tasks to queue
	task := bgjobs.NewTask(TypeBankDelete, []byte("goodbye"))
	task2 := bgjobs.NewTask(TypeBankAdd, []byte("hello"))
	task3 := bgjobs.NewTask(TypeBankUpdate, []byte("howdy"))
	if _, err := client.Enqueue(task); err != nil {
		log.Fatal(err)
	}
	if _, err := client.Enqueue(task2); err != nil {
		log.Fatal(err)
	}
	if _, err := client.Enqueue(task3); err != nil {
		log.Fatal(err)
	}

	// worker
	srv := bgjobs.NewServer(bgjobs.RedisOpts{Addr: Addr}, bgjobs.ServerConfig{})
	mux := bgjobs.NewServeMux()
	mux.HandleFunc(TypeBankDelete, testHandler)
	mux.HandleFunc(TypeBankAdd, testHandler)
	mux.HandleFunc(TypeBankUpdate, testHandler)

	srv.Run(mux)
}

func testHandler(t *bgjobs.Task) error {
	var buffer bytes.Buffer
	buffer.Write(t.Payload)
	fmt.Println(buffer.String())
	return nil
}
