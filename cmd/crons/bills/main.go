package main

import (
	"context"
	"log"

	"github.com/Chance093/roomie-bills/internal/lib/taskqueue"
	"github.com/Chance093/roomie-bills/internal/tasks"
)

// run as a cron job every saturday
func main() {
	// starts task pipeline for getting outstanding bills and then new bills
	tc := taskqueue.NewClient(context.Background(), taskqueue.ClientOpts{})

	newTask, err := tasks.NewGetOutstandingBillsTask()
	if err != nil {
		log.Fatalf("Could not create starting task for cron: %s", err.Error())
	}

	if _, err := tc.Enqueue(newTask); err != nil {
		log.Fatalf("Could not enqueue starting task: %s", err.Error())
	}
}
