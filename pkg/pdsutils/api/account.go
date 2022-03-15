// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Account struct comprise of Context , PDS api client.
type Account struct {
	Context   context.Context
	apiClient *pds.APIClient
}

// GetAccountsList function return list of Account objects.
func (account *Account) GetAccountsList() ([]pds.ModelsAccount, error) {
	client := account.apiClient.AccountsApi
	log.Info("Get list of Accounts.")
	accountsModel, res, err := client.ApiAccountsGet(account.Context).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Debugf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accountsModel.GetData(), nil
}

// GetAccount return an Account object.
func (account *Account) GetAccount(accountID string) (*pds.ModelsAccount, error) {
	client := account.apiClient.AccountsApi
	log.Infof("Get the account detail having UUID: %v", accountID)
	accountModel, res, err := client.ApiAccountsIdGet(account.Context, accountID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Full HTTP response: %v", res)
		log.Debugf("Error when calling `ApiAccountsGet``: %v", err)
		return nil, err
	}
	return accountModel, nil
}

// GetAccountUsers return list of user objects.
func (account *Account) GetAccountUsers(accountID string) ([]pds.ModelsUser, error) {
	log.Infof("Get users for the account having UUID: %v ", accountID)
	client := account.apiClient.AccountsApi
	accountInfo, er := account.GetAccount(accountID)
	if er != nil {
		return nil, er
	}
	log.Infof("Get the users belong to the account having name: %s", accountInfo.GetName())
	usersModel, res, err := client.ApiAccountsIdUsersGet(account.Context, accountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Full HTTP response: %v", res)
		log.Errorf("Error when calling `ApiAccountsGet``: %v", err)
		return nil, err
	}
	return usersModel.GetData(), nil
}

// CreateNewAccount create a new account and return the account object.
func (account *Account) CreateNewAccount(name string) (*pds.ModelsAccount, error) {
	log.Info("Create new account.")
	createAccountReq := pds.ControllersCreateAccountRequest{Name: &name}
	client := account.apiClient.AccountsApi
	accountModel, res, err := client.ApiAccountsPost(account.Context).Body(createAccountReq).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Full HTTP response: %v", res)
		log.Debugf("Error when calling `ApiAccountsPost``: %v\n", err)
		return nil, err
	}
	return accountModel, nil
}
