package main

import (
	"log"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

// run as a cron job every saturday
func main() {
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	// get access tokens from bank table in db

	// get transactions from plaid using access tokens
	pc := plaid.NewClient(env)
	pc.GetNewTransactions()

	// parse transactions for bills only

	// insert into db, ignoring already existing bills (plaid id is unique)

	// Send discord message notifying chat of new bills, or if no new bills have come in
}
