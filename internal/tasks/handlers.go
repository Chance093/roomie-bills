package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

type Handler struct {
	pc plaid.Client
	jc bgjobs.Client
	db *db.DB
}

func NewHandler(pc plaid.Client, jc bgjobs.Client, db *db.DB) Handler {
	return Handler{pc, jc, db}
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
