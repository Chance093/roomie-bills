package crons

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/types"
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
	plaidBills, err := c.pc.GetBills(ctx, accessTokens)
	if err != nil {
		return fmt.Errorf("Error while getting bills from plaid: %w", err)
	}

	// find new bills
	newPlaidBills, err := c.db.FindNewPlaidBills(plaidBills)
	if err != nil {
		return fmt.Errorf("Error while getting non existing bills: %w", err)
	}

	// if no new bills, send discord message notifying no new bills and end execution
	if len(newPlaidBills) == 0 {
		if err := c.dc.SendNoBillsMessage(); err != nil {
			return fmt.Errorf("Error while sending no bills message to discord: %w", err)
		}

		return nil
	}

	// insert payments for bill owners
	bills, err := c.db.AddBillsAndPayments(newPlaidBills)
	if err != nil {
		return fmt.Errorf("Error while adding owner payments; %w", err)
	}

	// split bills 4 ways
	splitBills := splitBills(bills)

	// Send discord message notifying chat of new bills
	if err := c.dc.SendBills(splitBills); err != nil {
		return fmt.Errorf("Error while sending bills to discord: %w", err)
	}

	return nil
}

// TODO: use decimal package
func splitBills(bills []db.Bill) []types.SplitBill {
	splitBills := make([]types.SplitBill, len(bills))
	for i, bill := range bills {
		split := math.Round(bill.Total/4*100) / 100
		splitBills[i] = types.SplitBill{Bill: bill, Split: split}
	}

	return splitBills
}

func (c Cron) EndOfMonthSummaryCron() {
	// check db for unpaid bills
	// send discord message
}
