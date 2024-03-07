package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

// GetTenantListV1
func GetTenantListV1(accountID string) ([]apiStructs.WorkFlowResponse, error) {
	tenList, err := v2Components.Platform.ListTenants(accountID)
	if err != nil {
		return nil, err
	}
	return tenList, nil
}
