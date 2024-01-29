package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// TenantV2 struct
type TenantV2 struct {
	apiClientV2 *pdsv2.APIClient
}

// GetTenantsList return pds tenants models.
func (tenant *TenantV2) GetTenantsList(accountID string) ([]pdsv2.V1Tenant, error) {
	tenantClient := tenant.apiClientV2.TenantServiceApi
	log.Info("Get list of tenants.")
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tenantsModel, res, err := tenantClient.TenantServiceListTenants2(ctx, accountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TenantServiceListTenants2`: %v\n.Full HTTP response: %v", err, res)
	}
	return tenantsModel.Tenants, nil
}

// GetTenant return tenant model.
func (tenant *TenantV2) GetTenant(tenantID string) (*pdsv2.V1Tenant, error) {
	tenantClient := tenant.apiClientV2.TenantServiceApi
	log.Info("Get tenant.")
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tenantModel, res, err := tenantClient.TenantServiceGetTenant(ctx, tenantID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TenantServiceGetTenant`: %v\n.Full HTTP response: %v", err, res)
	}
	return tenantModel, nil
}

//tenantClient.ApiTenantsIdDnsDetailsGet not available
