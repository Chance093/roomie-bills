package main

import (
	"log"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

func main() {
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	// get transactions from plaid
	pc := plaid.NewClient(env)
	pc.GetNewTransactions()

	// get transactions from db

	// get unaccounted transactions (what is not in db yet)

	// split bill 4 ways

	// send off discord message

	// save new transactions to db
}
