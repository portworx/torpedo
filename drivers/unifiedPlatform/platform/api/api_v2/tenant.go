package apiv2

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

// AccountV2 struct
type PLATFORM_API_V2 struct {
	ApiClientV2 *platformV2.APIClient
}

// GetClient updates the header with bearer token and returns the new client
func (AccountV2 *PLATFORM_API_V2) getClient() (context.Context, *platformV2.AccountServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V2 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	AccountV2.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	client := AccountV2.ApiClientV2.AccountServiceAPI

	return ctx, client, nil
}

// GetAccountList returns the list of accounts
func (AccountV2 *PLATFORM_API_V2) GetTenantList() {
	log.Infof("This is a call from tenant from PLATFORM_API_V2")
}
