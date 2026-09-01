package crons

import (
	"context"
	"fmt"
	"time"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

type Cron struct {
	pc plaid.Client
	dc lib.DiscordClient
	db *db.DB
}

func New(pc plaid.Client, dc lib.DiscordClient, db *db.DB) Cron {
	return Cron{pc, dc, db}
}

func (c Cron) CheckForNewBills() error {
	// get access tokens from bank table in db
	accessTokens, err := c.db.GetBankAccessTokens()
	if err != nil {
		return fmt.Errorf("Error while getting bank access tokens: %w", err)
	}

	// get transactions from plaid using access tokens
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	raw, err := c.pc.GetTransactions(ctx, accessTokens)
	if err != nil {
		return fmt.Errorf("Error while getting transactions from plaid: %w", err)
	}

	fmt.Println(raw)

	// parse transactions for bills only

	// insert into db, ignoring already existing bills (plaid id is unique)

	// Send discord message notifying chat of new bills, or if no new bills have come in
	return nil
}

func (c Cron) EndOfMonthSummaryCron() {
	// check db for unpaid bills
	// send discord message
}
