package main

import (
	"log"

	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/tasks"
)

// run as a cron job every saturday
func main() {
	// starts task pipeline for getting bills
	redisOpts := bgjobs.RedisOpts{}
	jc := bgjobs.NewClient(redisOpts)
	defer jc.Close()

	newTask, err := tasks.NewGetUnpaidBillsTask()
	if err != nil {
		log.Fatalf("Could not create starting task for cron: %s", err.Error())
	}

	if _, err := jc.Enqueue(newTask); err != nil {
		log.Fatalf("Could not enqueue starting task: %s", err.Error())
	}
}

// query db for unpaid bills
// send message for all unpaid bills
// need name of roomie who owes, name of roomie who is being paid, bill, and amount
