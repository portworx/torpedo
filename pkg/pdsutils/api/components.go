// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
)

// Components struct contain all the conponent of PDS to access all the assciated funcationality.
type Components struct {
	Account            *Account
	Tenant             *Tenant
	Project            *Project
	AccountRoleBinding *AccountRoleBinding
}

// NewComponents create an object of Components.
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
