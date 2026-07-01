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

type HostedLink struct {
	LinkToken string
	Url       string
	RequestId string
}

func (pc *PlaidClient) GetHostedLink(roomie string) (HostedLink, error) {
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

	// TODO: Set webhook url here
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})
	request.SetLinkCustomizationName("default")
	request.SetAccountFilters(accountFilters)
	request.SetHostedLink(hosted)
	request.SetUser(user)

	ctx := context.Background()
	linkTokenCreateResp, _, err := pc.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
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
