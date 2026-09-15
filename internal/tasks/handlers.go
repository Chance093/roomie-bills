package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/discord"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/types"
	"github.com/Chance093/roomie-bills/internal/utils"
)

type Handler struct {
	pc plaid.Client
	jc bgjobs.Client
	dc discord.Client
	db *db.DB
}

func NewHandler(pc plaid.Client, jc bgjobs.Client, dc discord.Client, db *db.DB) Handler {
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

func (h Handler) GetOutstandingBills(ctx context.Context, t bgjobs.Task) error {
	outstandingBills, err := h.db.GetOutstandingBills()
	if err != nil {
		return fmt.Errorf("Error getting unpaid bills: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewSendOutstandingBillsTask(outstandingBills)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) SendOutstandingBills(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload SendOutstandingBillsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// send outstanding bills
	if err := h.dc.SendOutstandingBills(payload.OutstandingBills); err != nil {
		return fmt.Errorf("Error sending outstanding bills: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewGetAccessTokensTask()
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
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

	// create new task and enqueue depending on if there are bills or not
	var newTask *bgjobs.Task
	if len(plaidBills) == 0 {
		newTask, err = NewSendNoNewBillsTask()
	} else {
		newTask, err = NewGetNewBillsTask(plaidBills)
	}
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
		newTask, err = NewSendNoNewBillsTask()
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

	// create new task and enqueue
	newTask, err := NewGetUnpaidBillIdsTask()
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) GetUnpaidBillIds(ctx context.Context, t bgjobs.Task) error {
	// get unpaid bill id's
	unpaidBillIds, err := h.db.GetUnpaidBillIds()
	if err != nil {
		return fmt.Errorf("Error while getting unpaid bill ids: %w", err)
	}

	// create new task and enqueue
	newTask, err := NewSetDiscordCommandsTask(unpaidBillIds)
	if err != nil {
		return err
	}

	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) SetDiscordCommands(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload SetDiscordCommandsPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// reset /paid command
	if err := h.dc.SetCommands(payload.UnpaidBillIds); err != nil {
		return fmt.Errorf("Failed to reset discord command: %w", err)
	}

	return nil
}

// TODO: use decimal package
func splitBills(bills []db.Bill) []types.SplitBill {
	splitBills := make([]types.SplitBill, len(bills))
	for i, bill := range bills {
		split := utils.SplitFourWay(bill.Total)
		splitBills[i] = types.SplitBill{Bill: bill, Split: split}
	}

	return splitBills
}

func (h Handler) SendNoNewBills(ctx context.Context, t bgjobs.Task) error {
	// send discord message letting them know there are no new bills
	if err := h.dc.SendNoNewBillsMessage(); err != nil {
		return fmt.Errorf("Error sending no new bills message: %w", err)
	}

	return nil
}
