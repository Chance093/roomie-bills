package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

// TODO: find better naming convention for these
const (
	TypeGetAccessToken   = "get:accessToken"
	TypeGetBank          = "get:bank"
	TypeUpdateBank       = "update:bank"
	TypeGetAccounts      = "get:accounts"
	TypeAddAccounts      = "add:accounts"
	TypeGetAccessTokens  = "get:accessTokens"
	TypeGetBills         = "get:bills"
	TypeGetNewBills      = "get:new:bills"
	TypeAddBillsPayments = "add:bills:payments"
	TypeSendBills        = "send:bills"
	TypeSendNoBills      = "send:nothing"
)

type (
	GetAccessTokenPayload struct {
		PublicToken string `json:"publicToken"`
		LinkToken   string `json:"linkToken"`
	}

	GetBankPayload struct {
		AccessToken plaid.AccessToken `json:"accessToken"`
		LinkToken   string            `json:"linkToken"`
	}

	UpdateBankPayload struct {
		GetBankPayload
		Bank string `json:"bank"`
	}

	GetAccountsPayload struct {
		BankId      int    `json:"bankId"`
		AccessToken string `json:"accessToken"`
	}

	AddAccountsPayload struct {
		BankId   int             `json:"bankId"`
		Accounts []plaid.Account `json:"accounts"`
	}

	GetBillsPayload struct {
		AccessTokens []string `json:"accessTokens"`
	}

	GetNewBillsPayload struct {
		PlaidBills []plaid.Bill `json:"plaidBills"`
	}

	AddBillsPaymentsPayload struct {
		GetNewBillsPayload
	}

	SendBillsPayload struct {
		Bills []db.Bill `json:"bills"`
	}
)

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

func NewGetAccessTokensTask() (*bgjobs.Task, error) {
	return newTask("", TypeGetAccessTokens)
}

func NewGetBillsTask(accessTokens []string) (*bgjobs.Task, error) {
	v := GetBillsPayload{
		AccessTokens: accessTokens,
	}

	return newTask(v, TypeGetBills)
}

func NewGetNewBillsTask(plaidBills []plaid.Bill) (*bgjobs.Task, error) {
	v := GetNewBillsPayload{
		PlaidBills: plaidBills,
	}

	return newTask(v, TypeGetNewBills)
}

func NewAddBillsPaymentsTask(plaidBills []plaid.Bill) (*bgjobs.Task, error) {
	v := AddBillsPaymentsPayload{
		GetNewBillsPayload{PlaidBills: plaidBills},
	}

  return newTask(v, TypeAddBillsPayments)
}

func NewSendBillsTask(bills []db.Bill) (*bgjobs.Task, error) {
  v := SendBillsPayload{
    Bills: bills,
  }

  return newTask(v, TypeSendBills)
}

func NewSendNoBillsTask() (*bgjobs.Task, error) {
  return newTask("", TypeSendNoBills)
}
