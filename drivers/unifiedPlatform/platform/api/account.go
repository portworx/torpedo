package api

import (
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

func (AccountV2 *AccountV2) GetAccountList() ([]platformV2.V1Account1, error) {
	client := AccountV2.ApiClientV2.AccountServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	accountList, _, err := client.AccountServiceListAccounts(ctx).Execute()
	log.Info("Get list of Accounts.")
	return accountList.Accounts, nil
}

// GetAccount return pds account model.
func (AccountV2 *AccountV2) GetAccount(accountID string) (*platformV2.V1Account1, error) {
	client := AccountV2.ApiClientV2.AccountServiceAPI
	log.Infof("Get the account detail having UUID: %v", accountID)
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	accountModel, res, err := client.AccountServiceGetAccount(ctx, accountID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceGetAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return accountModel, nil
}

// CreateAccount return pds account model.
func (AccountV2 *AccountV2) CreateAccount() (*platformV2.V1Account1, error) {
	client := AccountV2.ApiClientV2.AccountServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	accountModel, res, err := client.AccountServiceCreateAccount(ctx).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceCreateAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return accountModel, nil
}

// DeleteBackupLocation delete backup location and return status.
func (AccountV2 *AccountV2) DeleteBackupLocation(accountId string) (*status.Response, error) {
	client := AccountV2.ApiClientV2.AccountServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := client.AccountServiceDeleteAccount(ctx, accountId).Execute()
	if err != nil {
		return nil, fmt.Errorf("Error when calling `AccountServiceDeleteAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
