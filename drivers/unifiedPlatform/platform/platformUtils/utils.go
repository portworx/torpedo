package platformUtils

import (
	pdsdriver "github.com/portworx/torpedo/drivers/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
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
func CreatePlatformAccountV1(name, displayName, userMail string) (*platformv1.V1Account1, error) {
	acc, err := v2Components.Platform.AccountV2.CreateAccount(name, displayName, userMail)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// move to separate dir
func WhoAmIV1() (*platformv1.V1WhoAmIResponse, error) {
	whoAmIResp, err := v2Components.Platform.WhoAmI.WhoAmI()
	if err != nil {
		return nil, err
	}
	return whoAmIResp, nil
}

// move to separate dir
func GetPlatformTenantListV1(accountID string) ([]platformv1.V1Tenant, error) {
	tenantsList, err := v2Components.Platform.TenantV2.ListTenants(accountID)
	if err != nil {
		return nil, err
	}
	return tenantsList, nil
}

// move to separate dir
func GetPlatformAccountListV1() ([]platformv1.V1Account1, error) {
	accList, err := v2Components.Platform.AccountV2.GetAccountList()
	if err != nil {
		return nil, err
	}
	return accList, nil
}

// move to separate dir
func GetPlatformAccountID(accList []platformv1.V1Account1, accountName string) string {
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
