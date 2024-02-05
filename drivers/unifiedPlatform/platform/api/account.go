package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// AccountV2 struct
type AccountV2 struct {
	ApiClientV2 *platformV2.APIClient
}

// GetClient updates the header with bearer token and returns the new client
func (AccountV2 *AccountV2) GetClient() (context.Context, *platformV2.AccountServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	AccountV2.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	client := AccountV2.ApiClientV2.AccountServiceAPI

	return ctx, client, nil
}

// GetAccountList returns the list of accounts
func (AccountV2 *AccountV2) GetAccountList() ([]platformV2.V1Account1, error) {
	ctx, client, err := AccountV2.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	accountList, res, err := client.AccountServiceListAccounts(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceListAccounts`: %v\n.Full HTTP response: %v", err, res)
	}
	return accountList.Accounts, nil
}

// GetAccount return pds account model.
func (AccountV2 *AccountV2) GetAccount(accountID string) (*platformV2.V1Account1, error) {
	log.Infof("Get the account detail having UUID: %v", accountID)
	ctx, client, err := AccountV2.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	accountModel, res, err := client.AccountServiceGetAccount(ctx, accountID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceGetAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return accountModel, nil
}

// CreateAccount return pds account model.
func (AccountV2 *AccountV2) CreateAccount(accountName, displayName, userMail string) (*platformV2.V1Account1, error) {
	_, client, err := AccountV2.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	var createRequest platformV2.ApiAccountServiceCreateAccountRequest
	createRequest = createRequest.V1Account1(platformV2.V1Account1{
		Meta: &platformV2.V1Meta{
			Name: &accountName,
		},
		Config: &platformV2.V1Config6{
			UserEmail:   &userMail,
			DisplayName: &displayName,
		},
	})

	accountModel, res, err := client.AccountServiceCreateAccountExecute(createRequest)

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceCreateAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return accountModel, nil
}

// DeleteBackupLocation delete backup location and return status.
func (AccountV2 *AccountV2) DeleteBackupLocation(accountId string) (*status.Response, error) {
	ctx, client, err := AccountV2.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	_, res, err := client.AccountServiceDeleteAccount(ctx, accountId).Execute()
	if err != nil {
		return nil, fmt.Errorf("Error when calling `AccountServiceDeleteAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
