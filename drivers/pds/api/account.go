// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Account struct
type Account struct {
	Context   context.Context
	apiClient *pds.APIClient
}

// GetAccountsList func
func (account *Account) GetAccountsList() ([]pds.ModelsAccount, error) {
	client := account.apiClient.AccountsApi
	log.Info("Get list of Accounts.")
	accountsModel, res, err := client.ApiAccountsGet(account.Context).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accountsModel.GetData(), nil
}

// GetAccount func
func (account *Account) GetAccount(accountID string) (*pds.ModelsAccount, error) {
	client := account.apiClient.AccountsApi
	log.Info("Get the account detail having UUID: %v", accountID)
	accountModel, res, err := client.ApiAccountsIdGet(account.Context, accountID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Full HTTP response: %v\n", res)
		log.Errorf("Error when calling `ApiAccountsIdGet``: %v\n", err)
		return nil, err
	}
	return accountModel, nil
}

// GetAccountUsers func
func (account *Account) GetAccountUsers(accountID string) ([]pds.ModelsUser, error) {
	client := account.apiClient.AccountsApi
	accountInfo, _ := account.GetAccount(accountID)
	log.Info("Get the users belong to the account having name: %v", accountInfo.GetName())
	usersModel, res, err := client.ApiAccountsIdUsersGet(account.Context, accountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Full HTTP response: %v\n", res)
		log.Errorf("Error when calling `ApiAccountsIdUsersGet``: %v\n", err)
		return nil, err
	}
	return usersModel.GetData(), nil
}

// AcceptEULA func
func (account *Account) AcceptEULA(accountID string, eulaVersion string) error {
	client := account.apiClient.AccountsApi
	accountInfo, _ := account.GetAccount(accountID)
	log.Info("Get the users belong to the account having name: %v", accountInfo.GetName())
	updateRequest := pds.ControllersAcceptEULARequest{
		Version: &eulaVersion,
	}
	res, err := client.ApiAccountsIdEulaPut(account.Context, accountID).Body(updateRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Full HTTP response: %v\n", res)
		log.Errorf("Error when calling `ApiAccountsIdUsersGet``: %v\n", err)
	}
	return err
}
