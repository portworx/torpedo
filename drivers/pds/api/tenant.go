package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Tenant struct
type Tenant struct {
	context   context.Context
	apiClient *pds.APIClient
}

// GetTenantsList func
func (tenant *Tenant) GetTenantsList(accountID string) ([]pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get list of tenants.")
	tenantsModel, res, err := tenantClient.ApiAccountsIdTenantsGet(tenant.context, accountID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdTenantsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantsModel.GetData(), nil
}

// GetTenant func
func (tenant *Tenant) GetTenant(tenantID string) (*pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get tenant.")
	tenantModel, res, err := tenantClient.ApiTenantsIdGet(tenant.context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantModel, nil
}

// GetDNS func
func (tenant *Tenant) GetDNS(tenantID string) (*pds.ModelsDNSDetails, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get tenant.")
	tenantDNSModel, res, err := tenantClient.ApiTenantsIdDnsDetailsGet(tenant.context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantDNSModel, nil
}
