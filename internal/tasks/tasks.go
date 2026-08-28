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
	TypeGetAccounts    = "get:accounts"
	TypeAddAccounts    = "add:accounts"
)

type GetAccessTokenPayload struct {
	PublicToken string `json:"publicToken"`
	LinkToken   string `json:"linkToken"`
}

type GetBankPayload struct {
	AccessToken plaid.AccessToken `json:"accessToken"`
	LinkToken   string            `json:"linkToken"`
}

type UpdateBankPayload struct {
	GetBankPayload
	Bank string `json:"bank"`
}

type GetAccountsPayload struct {
	BankId      int
	AccessToken string
}

type AddAccountsPayload struct {
	BankId   int
	Accounts []plaid.Account
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

func NewGetAccountsTask(accessToken string, bankId int) (*bgjobs.Task, error) {
	v := GetAccountsPayload{
		AccessToken: accessToken,
		BankId:      bankId,
	}

	return newTask(v, TypeGetAccounts)
}

func NewAddAccountsTask(accounts []plaid.Account, bankId int) (*bgjobs.Task, error) {
	v := AddAccountsPayload{
		Accounts: accounts,
		BankId:   bankId,
	}

	return newTask(v, TypeAddAccounts)
}
