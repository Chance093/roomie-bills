package main

import (
	"log"

	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
)

const Addr = "127.0.0.1:6379"

func main() {
	// processor
	client := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})
	task := bgjobs.NewTask("delete:bank", []byte("goodbye"))
	if _, err := client.Enqueue(task); err != nil {
		log.Fatal(err)
	}
}
