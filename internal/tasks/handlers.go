package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/types"
)

type Handler struct {
	pc plaid.Client
	jc bgjobs.Client
	dc lib.DiscordClient
	db *db.DB
}

func NewHandler(pc plaid.Client, jc bgjobs.Client, dc lib.DiscordClient, db *db.DB) Handler {
	return Handler{pc, jc, dc, db}
}

func (h Handler) GetAccessToken(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload GetAccessTokenPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// get access token from plaid
	accessToken, err := h.pc.GetAccessToken(ctx, payload.PublicToken)
	if err != nil {
		return fmt.Errorf("Could not get access token: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewGetBankTask(accessToken, payload.LinkToken)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) GetBankName(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload GetBankPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// get bank name from plaid
	bank, err := h.pc.GetBankName(ctx, payload.AccessToken)
	if err != nil {
		return fmt.Errorf("Could not get bank name from plaid: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewUpdateBankTask(payload.AccessToken, payload.LinkToken, bank)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) UpdateBank(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload UpdateBankPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// update bank record (idempotent)
	bankId, err := h.db.UpdateBankRecord(payload.LinkToken, payload.Bank, payload.AccessToken)
	if err != nil {
		return fmt.Errorf("Could not update bank record: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewGetAccountsTask(payload.AccessToken.Token, bankId)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) GetAccounts(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload GetAccountsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// get accounts associated with bank
	accounts, err := h.pc.GetAccounts(ctx, payload.AccessToken)
	if err != nil {
		return fmt.Errorf("Could not get accounts from plaid: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewAddAccountsTask(accounts, payload.BankId)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) AddAccounts(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload AddAccountsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// add accounts (idempotent)
	if err := h.db.AddAccounts(payload.Accounts, payload.BankId); err != nil {
		return fmt.Errorf("Could not add accounts to db: %w", err)
	}

	return nil
}

func (h Handler) GetAccessTokens(ctx context.Context, t bgjobs.Task) error {
	// get access tokens from bank table in db
	accessTokens, err := h.db.GetBankAccessTokens()
	if err != nil {
		return fmt.Errorf("Error while getting bank access tokens: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewGetBillsTask(accessTokens)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) GetBills(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload GetBillsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// get bills from plaid using access tokens
	plaidBills, err := h.pc.GetBills(ctx, payload.AccessTokens)
	if err != nil {
		return fmt.Errorf("Error while getting bills from plaid: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewGetNewBillsTask(plaidBills)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) GetNewBills(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload GetNewBillsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// find new bills
	newPlaidBills, err := h.db.FindNewPlaidBills(payload.PlaidBills)
	if err != nil {
		return fmt.Errorf("Error while getting non existing bills: %w", err)
	}

	// create new task and enqueue depending on if there are bills or not
	var newTask *bgjobs.Task
	if len(newPlaidBills) == 0 {
		newTask, err = NewSendNoBillsTask()
	} else {
		newTask, err = NewAddBillsPaymentsTask(newPlaidBills)
	}
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) AddBillsAndPayments(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload AddBillsPaymentsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// insert payments for bill owners
	bills, err := h.db.AddBillsAndPayments(payload.PlaidBills)
	if err != nil {
		return fmt.Errorf("Error while adding bills and payments; %w", err)
	}

	// create new task and enqueue
	newTask, err := NewSendBillsTask(bills)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) SendBills(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload SendBillsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// split bills 4 ways
	splitBills := splitBills(payload.Bills)

	// Send discord message notifying chat of new bills
	if err := h.dc.SendBills(splitBills); err != nil {
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
