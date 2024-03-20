package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// GetTenantListV1
func GetTenantListV1() ([]automationModels.PlatformTenant, error) {
	tenList, err := v2Components.Platform.ListTenants()
	if err != nil {
		return nil, err
	}
	return tenList, nil
}

func GetDefaultTenantId(accountID string) (string, error) {
	var tenantId string
	tenantList, err := GetTenantListV1()
	log.FailOnError(err, "error while getting tenant list")
	for _, tenant := range tenantList {
		log.Infof("Available tenant's %s under the account id %s", *tenant.Meta.Name, accountID)
		tenantId = *tenant.Meta.Uid
		break
	}
	return tenantId, nil
}
