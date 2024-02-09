package platformUtils

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	err          error
)

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	//v2Components, err = pdsdriver.InitUnifiedPlatformApiComponents(controlPlaneURL, accountID, false)

	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

// move to separate dir
func GetAccountListv1() ([]apiStructs.Account, error) {
	accList, err := v2Components.Platform.GetAccountList()
	if err != nil {
		return nil, err
	}
	return accList, nil
}

//func CreatePlatformAccountV1(name, displayName, userMail string) (apiStructs.Account, error) {
//	acc, _, err := v2Components.Platform.CreateAccount(name, displayName, userMail)
//	if err != nil {
//		return acc, err
//	}
//	return acc, nil
//}
