package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetClient updates the header with bearer token and returns the new client
func (tenant *PLATFORM_API_V1) GetClient() (context.Context, *platformv1.TenantServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	tenant.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	tenant.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = tenant.AccountID
	client := tenant.ApiClientV1.TenantServiceAPI

	return ctx, client, nil
}

// ListTenants return pds tenants models.
func (tenant *PLATFORM_API_V1) ListTenants(accountID string) ([]ApiResponse, error) {
	tenantsResponse := []ApiResponse{}
	ctx, tenantClient, err := tenant.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	tenantsModel, res, err := tenantClient.TenantServiceListTenants2(ctx, accountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TenantServiceListTenants2`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of tenants - [%v]", tenantsModel)
	copier.Copy(&tenantsResponse, tenantsModel.Tenants)
	log.Infof("Value of accounts after copy - [%v]", tenantsResponse)

	return tenantsResponse, nil
}
