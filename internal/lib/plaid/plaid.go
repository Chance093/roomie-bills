package plaid

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/plaid/plaid-go/v43/plaid"
)

// TODO: create documentation for this package

type Client struct {
	client *plaid.APIClient
	env    map[string]string
}

func NewClient(env map[string]string) Client {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", env["PLAID_CLIENT_ID"])
	configuration.AddDefaultHeader("PLAID-SECRET", env["PLAID_SANDBOX_SECRET"])
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)

	return Client{client, env}
}

type HostedLink struct {
	LinkToken string
	Url       string
	RequestId string
}

func (c Client) GetHostedLink(ctx context.Context, roomie string) (HostedLink, error) {
	// init hosted link config
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: roomie,
	}
	depository := plaid.DepositoryFilter{
		AccountSubtypes: []plaid.DepositoryAccountSubtype{
			plaid.DEPOSITORYACCOUNTSUBTYPE_CHECKING,
		},
	}
	credit := plaid.CreditFilter{
		AccountSubtypes: []plaid.CreditAccountSubtype{plaid.CREDITACCOUNTSUBTYPE_CREDIT_CARD},
	}
	accountFilters := plaid.LinkTokenAccountFilters{
		Depository: &depository,
		Credit:     &credit,
	}
	request := plaid.NewLinkTokenCreateRequest(
		"Roomie Bills",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
	)
	hosted := plaid.LinkTokenCreateHostedLink{}
	webhookUrl := fmt.Sprintf("%s/webhooks/plaid", c.env["DOMAIN"])

	// set config in request
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})
	request.SetLinkCustomizationName("default")
	request.SetWebhook(webhookUrl)
	request.SetAccountFilters(accountFilters)
	request.SetHostedLink(hosted)
	request.SetUser(user)

	// send request to plaid
	linkTokenCreateResp, _, err := c.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return HostedLink{}, err
	}
	linkToken := linkTokenCreateResp.GetLinkToken()
	hostedLink := linkTokenCreateResp.GetHostedLinkUrl()
	requestId := linkTokenCreateResp.GetRequestId()

	return HostedLink{
		LinkToken: linkToken,
		Url:       hostedLink,
		RequestId: requestId,
	}, nil
}

type AccessToken struct {
	Token  string `json:"token"`
	ItemId string `json:"itemId"`
}

func (c Client) GetAccessToken(ctx context.Context, publicToken string) (AccessToken, error) {
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, _, err := c.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		return AccessToken{}, err
	}

	accessToken := exchangePublicTokenResp.GetAccessToken()
	itemId := exchangePublicTokenResp.GetItemId()

	return AccessToken{
		Token:  accessToken,
		ItemId: itemId,
	}, nil
}

func (c Client) GetBankName(ctx context.Context, accessToken AccessToken) (string, error) {
	request := plaid.NewItemGetRequest(accessToken.Token)
	resp, _, err := c.client.PlaidApi.ItemGet(ctx).ItemGetRequest(*request).Execute()
	if err != nil {
		return "", err
	}

	item := resp.GetItem()
	institution := item.GetInstitutionName()

	return institution, nil
}

type Account struct {
	PlaidId string
	Name    string
	Type    string
}

func (c Client) GetAccounts(ctx context.Context, accessToken string) ([]Account, error) {
	accountsGetRequest := plaid.NewAccountsGetRequest(accessToken)

	accountsGetResp, _, err := c.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*accountsGetRequest,
	).Execute()
	if err != nil {
		return nil, err
	}

	// parse through accounts and get relevant data
	rawAccs := accountsGetResp.GetAccounts()

	accounts := make([]Account, len(rawAccs))
	for i, acc := range rawAccs {
		plaidId := acc.GetAccountId()
		name := acc.GetName()
		accType := string(acc.GetSubtype())

		accounts[i] = Account{plaidId, name, accType}
	}

	return accounts, err
}

type Transaction struct {
	PlaidId string
	Payee   string
	Date    string
	Total   float64 // TODO: change to decimal package
}

type AccountTransactions struct {
	PlaidId      string
	Name         string
	Transactions []Transaction
}

func (c Client) GetTransactions(ctx context.Context, accessTokens []string) ([]AccountTransactions, error) {
	const iso8601TimeFormat = "2006-01-02"
	startDate := time.Now().Add(-7 * 24 * time.Hour).Format(iso8601TimeFormat)
	endDate := time.Now().Format(iso8601TimeFormat)

	// OPTIMIZE: This is a brute force solution
	var accountTransactions []AccountTransactions
	for _, at := range accessTokens {
		request := plaid.NewTransactionsGetRequest(
			at,
			startDate,
			endDate,
		)

		options := plaid.NewTransactionsGetRequestOptions()
		options.SetCount(100)
		options.SetOffset(0)

		request.SetOptions(*options)

		transactionsResp, _, err := c.client.PlaidApi.TransactionsGet(ctx).TransactionsGetRequest(*request).Execute()
		if err != nil {
			return nil, err // TODO: better error handling
		}

		// set accounts
		rawAccs := transactionsResp.GetAccounts()
		for _, rawAcc := range rawAccs {
			accountTransactions = append(accountTransactions, AccountTransactions{
				PlaidId:      rawAcc.AccountId,
				Name:         rawAcc.Name,
				Transactions: make([]Transaction, 0),
			})
		}

		// set transactions in accounts
		rawTransactions := transactionsResp.GetTransactions()
		// loop through each plaid transaction
		for _, rawTrans := range rawTransactions {
			// find index of account associated with transaction
			i := -1
			for j, accTrans := range accountTransactions {
				if rawTrans.GetAccountId() == accTrans.PlaidId {
					i = j
					break
				}
			}

			if i == -1 {
				return nil, errors.New("Could not find account!!! Should not happen")
			}

			// append transaction to that accounts transactions
			accountTransactions[i].Transactions = append(accountTransactions[i].Transactions, Transaction{
				PlaidId: rawTrans.GetTransactionId(),
				Payee:   rawTrans.GetName(),
				Date:    rawTrans.GetDate(),
				Total:   rawTrans.Amount, // TODO: change to decimal package
			})
		}
	}

	return accountTransactions, nil
}
