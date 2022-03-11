// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/sirupsen/logrus"
)

// Tenant struct comprise of Context , PDS api client.
type Tenant struct {
	Context   context.Context
	apiClient *pds.APIClient
}

// GetTenantsList return list of tenant objects.
func (tenant *Tenant) GetTenantsList(accountID string) ([]pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	logrus.Info("Get list of Tenants.")
	tenantsModel, res, err := tenantClient.ApiAccountsIdTenantsGet(tenant.Context, accountID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		logrus.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		logrus.Debugf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantsModel.GetData(), nil
}

// GetTenant return tenant object.
func (tenant *Tenant) GetTenant(tenantID string) (*pds.ModelsTenant, error) {
	tenantClient := tenant.apiClient.TenantsApi
	logrus.Info("Get the tenant.")
	tenantModel, res, err := tenantClient.ApiTenantsIdGet(tenant.Context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		logrus.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		logrus.Debugf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return tenantModel, nil
}
