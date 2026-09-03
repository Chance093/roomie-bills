package crons

import (
	"context"
	"fmt"
	"math"
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

	// get bills from plaid using access tokens
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bills, err := c.pc.GetBills(ctx, accessTokens)
	if err != nil {
		return fmt.Errorf("Error while getting bills from plaid: %w", err)
	}

	// TODO: Make these 2 db inserts a transaction (idempotent)

	// insert into db, ignoring already existing bills (plaid id is unique)
	newBills, err := c.db.AddBills(bills)
	if err != nil {
		return fmt.Errorf("Error while adding bills to db: %w", err)
	}

	if len(newBills) > 0 {
		// insert payments for bill owners
		if err := c.db.AddOwnerPayments(newBills); err != nil {
			return fmt.Errorf("Error while adding owner payments; %w", err)
		}

		// split bills 4 ways
		splitBills := splitBills(newBills)
		for _, bill := range splitBills {
			fmt.Println(bill)
		}

		// Send discord message notifying chat of new bills
	} else {
		// send discord message notifying no new bills this week
	}

	return nil
}

type SplitBill struct {
	plaid.Bill
	Split float64
}

// TODO: use decimal package
func splitBills(bills []plaid.Bill) []SplitBill {
	splitBills := make([]SplitBill, len(bills))
	for i, bill := range bills {
		split := math.Round(bill.Total/4*100) / 100
		splitBills[i] = SplitBill{bill, split}
	}

	return splitBills
}

func (c Cron) EndOfMonthSummaryCron() {
	// check db for unpaid bills
	// send discord message
}
