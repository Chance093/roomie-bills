package lib

import (
	"fmt"
	"github.com/plaid/plaid-go/v43/plaid"
)

type PlaidClient struct {
	client *plaid.APIClient
}

func NewPlaidClient(id, secret string) PlaidClient {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", id) 
	configuration.AddDefaultHeader("PLAID-SECRET", secret) 
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)
	return PlaidClient{client}
}

func (pc *PlaidClient) GetNewTransactions() {
	pc.getAllRecentTransactions()
	fmt.Println("getting past transactions")
}

func (pc *PlaidClient) getAllRecentTransactions() {
	// request from plaid api
	fmt.Println("getting all transactions")
}

func filterTransactions() {
	// find all new transactions
}
