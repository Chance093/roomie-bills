package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

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
	AccessToken plaid.AccessToken `json:"accessToken"`
	LinkToken   string          `json:"linkToken"`
}

type UpdateBankPayload struct {
	GetBankPayload
	Bank string `json:"bank"`
}

func newTask(v any, taskType string) (*bgjobs.Task, error) {
	p, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal struct into json: %w\n", err)
	}

	return bgjobs.NewTask(taskType, p), nil
}

func NewGetAccessTokenTask(publicToken, linkToken string) (*bgjobs.Task, error) {
	v := GetAccessTokenPayload{
		PublicToken: publicToken,
		LinkToken:   linkToken,
	}

	return newTask(v, TypeGetAccessToken)
}

func NewGetBankTask(accessToken plaid.AccessToken, linkToken string) (*bgjobs.Task, error) {
	v := GetBankPayload{
		AccessToken: accessToken,
		LinkToken:   linkToken,
	}

	return newTask(v, TypeGetBank)
}

func NewUpdateBankTask(accessToken plaid.AccessToken, linkToken, bank string) (*bgjobs.Task, error) {
	v := UpdateBankPayload{
		GetBankPayload: GetBankPayload{
			AccessToken: accessToken,
			LinkToken:   linkToken,
		},
		Bank: bank,
	}

	return newTask(v, TypeUpdateBank)
}
