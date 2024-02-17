package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	err          error
)

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

// GetAccountListv1
func GetAccountListv1() ([]apiStructs.WorkFlowResponse, error) {
	accList, err := v2Components.Platform.GetAccountList()
	if err != nil {
		return nil, err
	}
	return accList, nil
}

// GetAccount
func GetAccount(accountID string) (*apiStructs.WorkFlowResponse, error) {
	acc, err := v2Components.Platform.GetAccount(accountID)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// GetPlatformAccountID
func GetPlatformAccountID(accList []apiStructs.WorkFlowResponse, accountName string) string {
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

// CreatePlatformAccountV1
func CreatePlatformAccountV1(name, displayName, userMail string) (apiStructs.WorkFlowResponse, error) {
	acc, err := v2Components.Platform.CreateAccount(name, displayName, userMail)
	if err != nil {
		return acc, err
	}
	return acc, nil
}
