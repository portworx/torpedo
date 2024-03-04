package api

import (
	"context"
	"fmt"
	status "net/http"

	"github.com/jinzhu/copier"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

// AccountV1 struct
type PLATFORM_API_V1 struct {
	ApiClientV1 *platformv1.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (account *PLATFORM_API_V1) getAccountClient() (context.Context, *platformv1.AccountServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	account.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	account.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = account.AccountID

	client := account.ApiClientV1.AccountServiceAPI
	return ctx, client, nil
}

// GetAccountList returns the list of accounts
func (AccountV1 *PLATFORM_API_V1) GetAccountList() ([]WorkFlowResponse, error) {
	ctx, client, err := AccountV1.getAccountClient()
	accountsResponse := []WorkFlowResponse{}

	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	var getRequest platformv1.ApiAccountServiceListAccountsRequest
	getRequest = getRequest.ApiService.AccountServiceListAccounts(ctx)
	accountList, res, err := client.AccountServiceListAccountsExecute(getRequest)

	//accountList, res, err := client.AccountServiceListAccounts(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceListAccounts`: %v\n.Full HTTP response: %v", err, res)
	}

	log.Infof("Accounts - [%v]", accountList)

	log.Infof("Value of accounts - [%+v]", accountList.Accounts[0].Meta.Name)
	err = copier.Copy(&accountsResponse, accountList.Accounts)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of accounts after copy - [%v]", accountsResponse)

	return accountsResponse, nil
}

// GetAccount return pds account model.
func (AccountV1 *PLATFORM_API_V1) GetAccount(accountID string) (*WorkFlowResponse, error) {
	log.Infof("Get the account detail having UUID: %v", accountID)
	accountResponse := WorkFlowResponse{}
	ctx, client, err := AccountV1.getAccountClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	var getRequest platformv1.ApiAccountServiceGetAccountRequest
	getRequest = getRequest.ApiService.AccountServiceGetAccount(ctx, accountID)
	accountModel, res, err := client.AccountServiceGetAccountExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceGetAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of account - [%v]", *accountModel.Meta.Name)
	err = copier.Copy(&accountResponse, accountModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of account after copy - [%v]", *accountResponse.Meta.Name)
	return &accountResponse, nil
}

// CreateAccount return pds account model.
func (AccountV1 *PLATFORM_API_V1) CreateAccount(accountName, displayName, userMail string) (WorkFlowResponse, error) {
	_, client, err := AccountV1.getAccountClient()
	accountResponse := WorkFlowResponse{}
	if err != nil {
		return accountResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var createRequest platformv1.ApiAccountServiceCreateAccountRequest
	createRequest = createRequest.V1Account1(platformv1.V1Account1{
		Meta: &platformv1.V1Meta{
			Name: &accountName,
		},
		Config: &platformv1.V1Config6{
			UserEmail:   &userMail,
			DisplayName: &displayName,
		},
	})
	accountModel, res, err := client.AccountServiceCreateAccountExecute(createRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return accountResponse, fmt.Errorf("Error when calling `AccountServiceCreateAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of account - [%v]", accountResponse)
	err = copier.Copy(&accountResponse, accountModel)
	if err != nil {
		return WorkFlowResponse{}, err
	}
	log.Infof("Value of account after copy - [%v]", accountResponse)
	return accountResponse, nil
}

// DeleteBackupLocation delete backup location and return status.
func (AccountV1 *PLATFORM_API_V1) DeleteAccount(accountId string) error {
	ctx, client, err := AccountV1.getAccountClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	_, res, err := client.AccountServiceDeleteAccount(ctx, accountId).Execute()
	if err != nil {
		return fmt.Errorf("Error when calling `AccountServiceDeleteAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}
