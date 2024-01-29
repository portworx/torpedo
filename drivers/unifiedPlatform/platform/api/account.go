package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// AccountV2 struct
type AccountV2 struct {
	apiClientV2 *pdsv2.APIClient
}

func (account *AccountV2) GetAccountList() ([]pdsv2.V1Account, error) {
	client := account.apiClientV2.AccountServiceApi
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	log.Info("Get list of Accounts.")
	accountList, _, err := client.AccountServiceListAccounts(ctx).Execute()

	if err != nil {
		return nil, fmt.Errorf("Error when calling `AccountServiceListAccounts`: %v\n", err)
	}
	return accountList.Accounts, nil
}

// GetAccount return pds account model.
func (account *AccountV2) GetAccount(accountID string) (*pdsv2.V1Account, error) {
	client := account.apiClientV2.AccountServiceApi
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

//client.ApiAccountsIdUsersGet api not available

//client.ApiAccountsIdEulaPut api not available
