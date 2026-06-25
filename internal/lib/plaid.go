package lib

import "fmt"

type PlaidClient struct {
	CLIENT_ID string
	SECRET    string
}

func NewPlaidClient(id, secret string) PlaidClient {
	return PlaidClient{CLIENT_ID: id, SECRET: secret}
}

func (pc *PlaidClient) GetPastTransactions() {
	fmt.Println("getting past transactions")
}
