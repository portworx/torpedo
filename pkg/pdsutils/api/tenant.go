// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

/*
	Tenant struct consists the context and apiclient.
	It can be used to call the CRUD functionality with respect to Tenant API.
*/
type Tenant struct {
	Context   context.Context
	apiClient *pds.APIClient
}

/*
	Get the List of all the Tenants.
	@return []pds.ModelsTenant, error
*/
func (tenant *Tenant) GetTenantsList(accountId string) ([]pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get list of Accounts.")
	tenantsModel, res, err := tenantClient.ApiAccountsIdTenantsGet(tenant.Context, accountId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantsModel.GetData(), nil
}

/*
	Get the Tenant details.
	@param accountId string - Tenant UUID.
	@return *pds.ModelsTenant, error
*/
func (tenant *Tenant) GetTenant(tenantId string) (*pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	log.Info("Get list of Accounts.")
	tenantModel, res, err := tenantClient.ApiTenantsIdGet(tenant.Context, tenantId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantModel, nil
}
