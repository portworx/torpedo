package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// GetTenantListV1
func GetTenantListV1(accountID string) ([]apiStructs.WorkFlowResponse, error) {
	tenList, err := v2Components.Platform.ListTenants(accountID)
	if err != nil {
		return nil, err
	}
	return tenList, nil
}

func GetTenantId(accountId string) (string, error) {
	var tenantId string
	tenantList, err := GetTenantListV1(accountId)
	if err != nil {
		return "", err
	}
	for _, tenant := range tenantList {
		log.Infof("Available tenant's %s under the account id %s", *tenant.Meta.Name, accountId)
		tenantId = *tenant.Meta.Uid
		break
	}
	log.Infof("TenantID [%s]", tenantId)
	return tenantId, nil
}
