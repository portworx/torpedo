package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedControlPlane/utils"
	"github.com/portworx/torpedo/pkg/log"
)

// Account struct
type Accountv2 struct {
	ApiClientv2 *pdsv2.APIClient
}

func (accountv2 *Accountv2) GetAccountList() (*pdsv2.V1ListAccountsResponse, error) {
	client := accountv2.ApiClientv2.AccountServiceApi
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	accountList, _, err := client.AccountServiceListAccounts(ctx).Execute()
	log.Info("Get list of Accounts.")
	return accountList, nil
}
