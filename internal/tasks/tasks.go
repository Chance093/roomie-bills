package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
)

type Handler struct {
	pc lib.PlaidClient
	jc bgjobs.Client
	db *db.DB
}

func NewHandler(pc lib.PlaidClient, jc bgjobs.Client, db *db.DB) Handler {
	return Handler{pc, jc, db}
}

const (
	TypeGetAccessToken = "get:accessToken"
	TypeGetBank        = "get:bank"
	TypeUpdateBank     = "update:bank"
)

type GetAccessTokenPayload struct {
	PublicToken string `json:"publicToken"`
	LinkToken   string `json:"linkToken"`
}

type GetBankPayload struct {
	AccessToken lib.AccessToken `json:"accessToken"`
	LinkToken   string          `json:"linkToken"`
}

type UpdateBankRecordPayload struct {
	GetBankPayload
	Bank string `json:"bank"`
}

func (h Handler) GetAccessToken(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload GetAccessTokenPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// get access token from plaid
	accessToken, err := h.pc.GetAccessToken(payload.PublicToken)
	if err != nil {
		return fmt.Errorf("Could not get access token: %w", err)
	}

	// marshal new payload into json
	newPayload, err := json.Marshal(GetBankPayload{accessToken, payload.LinkToken})
	if err != nil {
		return fmt.Errorf("Could not marshal struct into json: %w", err)
	}

	// enqueue new task
	newTask := bgjobs.NewTask(TypeGetBank, newPayload)
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
	bank, err := h.pc.GetBankName(payload.AccessToken)
	if err != nil {
		return fmt.Errorf("Could not get bank name from plaid: %w", err)
	}

	// marshal new payload into json
	newPayload, err := json.Marshal(UpdateBankRecordPayload{
		GetBankPayload: GetBankPayload{
			AccessToken: payload.AccessToken,
			LinkToken:   payload.LinkToken,
		},
		Bank: bank,
	})
	if err != nil {
		return fmt.Errorf("Could not marshal struct into json: %w", err)
	}

	// enqueue new task
	newTask := bgjobs.NewTask(TypeUpdateBank, newPayload)
	if _, err := h.jc.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

func (h Handler) UpdateBank(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var payload UpdateBankRecordPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// check if bank record has already been updated (idempotent)
	if exists, err := h.db.AccessTokenExists(payload.AccessToken.Token); err != nil {
		return fmt.Errorf("Error while checking if access token already exists in db: %w", err)
	} else if exists {
		return nil // access token exists and bank record has already been updated
	}

	// update bank record
	if err := h.db.UpdateBankRecord(payload.LinkToken, payload.Bank, payload.AccessToken); err != nil {
		return fmt.Errorf("Could not update bank record: %w", err)
	}

	return nil
}
