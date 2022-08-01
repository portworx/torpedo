package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type Tenant struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (tenant *Tenant) GetTenantsList(accountId string) ([]pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get list of tenants.")
	tenantsModel, res, err := tenantClient.ApiAccountsIdTenantsGet(tenant.context, accountId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdTenantsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantsModel.GetData(), nil
}

func (tenant *Tenant) GetTenant(tenantId string) (*pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get tenant.")
	tenantModel, res, err := tenantClient.ApiTenantsIdGet(tenant.context, tenantId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantModel, nil
}

func (tenant *Tenant) GetDns(tenantId string) (*pds.ModelsDNSDetails, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get tenant.")
	tenantDnsModel, res, err := tenantClient.ApiTenantsIdDnsDetailsGet(tenant.context, tenantId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantDnsModel, nil
}
