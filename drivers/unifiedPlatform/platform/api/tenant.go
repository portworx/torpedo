package api

import (
	"fmt"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// TenantV2 struct
type TenantV2 struct {
	ApiClientV2 *platformV2.APIClient
}

// ListTenants return pds tenants models.
func (tenant *TenantV2) ListTenants(accountID string) ([]platformV2.V1Tenant, error) {
	tenantClient := tenant.ApiClientV2.TenantServiceAPI
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
func (tenant *TenantV2) GetTenant(tenantID string) (*platformV2.V1Tenant, error) {
	tenantClient := tenant.ApiClientV2.TenantServiceAPI
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

// CreateTenant return tenant model.
func (tenant *TenantV2) CreateTenant() (*platformV2.V1Tenant, error) {
	tenantClient := tenant.ApiClientV2.TenantServiceAPI
	log.Info("Get tenant.")
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tenantModel, res, err := tenantClient.TenantServiceCreateTenant(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TenantServiceCreateTenant`: %v\n.Full HTTP response: %v", err, res)
	}
	return tenantModel, nil
}

// DeleteTenant return tenant model.
func (tenant *TenantV2) DeleteTenant(tenantId string) (*status.Response, error) {
	tenantClient := tenant.ApiClientV2.TenantServiceAPI
	log.Info("Get tenant.")
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := tenantClient.TenantServiceDeleteTenant(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return res, fmt.Errorf("Error when calling `TenantServiceDeleteTenant`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}

//tenantClient.ApiTenantsIdDnsDetailsGet not available
