package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// GetAccountListv1
//func GetAccountListv1() ([]automationModels.WorkFlowResponse, error) {
//	accList, err := v2Components.Platform.GetAccountList()
//	if err != nil {
//		return nil, err
//	}
//	return accList, nil
//}

// GetAccount
func GetAccount(accountID string) (*automationModels.WorkFlowResponse, error) {
	request := automationModels.PlatformAccount{
		Get: automationModels.PlatformGetAccount{
			AccountId: accountID,
		},
	}
	acc, err := v2Components.Platform.GetAccount(&request)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// GetPlatformAccountID
func GetPlatformAccountID(accList []automationModels.WorkFlowResponse, accountName string) string {
	var accID string
	for _, acc := range accList {
		if *acc.Meta.Name == accountName {
			log.Infof("Available account %s", *acc.Meta.Name)
			accID = *acc.Meta.Uid
			log.Infof("Available account ID %s", accID)
		}
	}
	return accID
}
