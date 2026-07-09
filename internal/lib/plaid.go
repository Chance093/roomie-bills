package lib

import (
	"context"
	"fmt"

	"github.com/plaid/plaid-go/v43/plaid"
)

type PlaidClient struct {
	client *plaid.APIClient
	ctx    context.Context
}

func NewPlaidClient(id, secret string) PlaidClient {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", id)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)
	ctx := context.Background()

	return PlaidClient{client, ctx}
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

	linkTokenCreateResp, _, err := pc.client.PlaidApi.LinkTokenCreate(pc.ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return HostedLink{}, err
	}

	linkToken := linkTokenCreateResp.GetLinkToken()
	hostedLink := linkTokenCreateResp.GetHostedLinkUrl()
	requestId := linkTokenCreateResp.GetRequestId()

	fmt.Printf("Hosted link obtained: %s\n", hostedLink)

	return HostedLink{
		LinkToken: linkToken,
		Url:       hostedLink,
		RequestId: requestId,
	}, nil
}

var jwkCache = map[string]*plaid.JWKPublicKey{}

func (pc *PlaidClient) GetJWK(kid string) (*plaid.JWKPublicKey, error) {
	if key, ok := jwkCache[kid]; ok && key != nil {
		return key, nil
	}
	req := *plaid.NewWebhookVerificationKeyGetRequest(kid)
	resp, _, err := pc.client.PlaidApi.WebhookVerificationKeyGet(pc.ctx).
		WebhookVerificationKeyGetRequest(req).
		Execute()
	if err != nil {
		return nil, err
	}
	key := resp.GetKey()
	if key.Kid == kid {
		jwkCache[kid] = &key
	}
	return &key, nil
}
