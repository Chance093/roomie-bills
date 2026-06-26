package lib

import "fmt"

type PlaidClient struct {
	CLIENT_ID string
	SECRET    string
}

func NewPlaidClient(id, secret string) PlaidClient {
	return PlaidClient{CLIENT_ID: id, SECRET: secret}
}

func (pc *PlaidClient) GetNewTransactions() {
	pc.getAllTransactions()
	fmt.Println("getting past transactions")
}

func (pc *PlaidClient) getAllTransactions() {
	// request from plaid api
	fmt.Println("getting all transactions")
	// filter for new transactions
	filterTransactions()
}

func filterTransactions() {
	// find all new transactions
}


