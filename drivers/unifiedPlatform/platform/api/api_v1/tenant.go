package apiv1

import (
	"github.com/portworx/torpedo/pkg/log"
)

// GetAccountList returns the list of accounts
func (AccountV2 *PLATFORM_API_V1) GetTenantList() {
	log.Infof("This is a call from tenant from PLATFORM_API_V1")
}
