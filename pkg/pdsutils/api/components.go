// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
)

/*
	Components struct comprise of all the component stuct to
	leverage the usage of the functionality and act as entry point.
*/
type Components struct {
	Account            *Account
	Tenant             *Tenant
	Project            *Project
	AccountRoleBinding *AccountRoleBinding
}

/*
	NewComponents create a struct literal that can be leveraged to call all the components functions.

	@param ctx context.Context - Context for authentication api request for the components.
	@param apiClient *pds.APIClient - PDS api client to invoke API request.
	@return *Component
*/
func NewComponents(Context context.Context, apiClient *pds.APIClient) *Components {
	return &Components{
		Account: &Account{
			Context:   Context,
			apiClient: apiClient,
		},
		Tenant: &Tenant{
			Context:   Context,
			apiClient: apiClient,
		},
		Project: &Project{
			Context:   Context,
			apiClient: apiClient,
		},
		AccountRoleBinding: &AccountRoleBinding{
			Context:   Context,
			apiClient: apiClient,
		},
	}
}
