package grpc

import (
	"github.com/portworx/torpedo/pkg/log"
)

// GetAccountList returns the list of accounts
func (AccountV2 *PLATFORM_GRPC) GetTenantList() {
	log.Infof("This is a call from tenant from PLATFORM_GRPC")
}
