package api

import (
	"fmt"
	//pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

// Account struct
type Accountv2 struct {
	ApiClientv2 *platformv2.APIClient
}

func (accountv2 *Accountv2) GetAccountList() (*platformv2.V1ListAccountsResponse, error) {
	client := accountv2.ApiClientv2.AccountServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	accountList, _, err := client.AccountServiceListAccounts(ctx).Execute()
	log.Info("Get list of Accounts.")
	return accountList, nil
}
