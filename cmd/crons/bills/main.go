package main

import (
	"log"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/crons"
	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

// run as a cron job every saturday
func main() {
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	pc := plaid.NewClient(env)
	dc, err := lib.NewDiscordClient(env)
	if err != nil {
		log.Fatal(err)
	}

	db := db.NewDB()
	defer db.Close()

	cron := crons.New(pc, dc, db)

	if err := cron.CheckForNewBills(); err != nil {
		log.Fatal(err)
	}
}
