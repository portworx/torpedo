package platformUtils

import (
	pdsdriver "github.com/portworx/torpedo/drivers/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	err          error
)

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = pdsdriver.InitUnifiedPlatformApiComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

// move to separate dir
func CreatePlatformAccountV1(name, displayName, userMail string) (apiStructs.Account, error) {
	acc, _, err := v2Components.Platform.CreateAccount(name, displayName, userMail)
	if err != nil {
		return acc, err
	}
	return acc, nil
}
