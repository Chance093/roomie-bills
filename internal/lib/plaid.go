package lib

import (
	"context"

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
}

func (pc *PlaidClient) GetAccessToken(userName string) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userName,
	}
	transactions := plaid.LinkTokenTransactions{}
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
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})
	request.SetTransactions(transactions)
	request.SetWebhook("https://sample-web-hook.com")
	request.SetRedirectUri("https://domainname.com/oauth-page.html")
	request.SetAccountFilters(accountFilters)

	ctx := context.Background()
	linkTokenCreateResp, _, err := pc.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		panic(err)
	}
	linkToken := linkTokenCreateResp.GetLinkToken()

	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(linkToken)
	exchangePublicTokenResp, _, err := pc.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	accessToken := exchangePublicTokenResp.GetAccessToken()

	// TODO: Save access token to db
}
