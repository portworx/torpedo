package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// ServiceAccount struct
type ServiceAccount struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListServiceAccounts func
func (sa *ServiceAccount) ListServiceAccounts(tenantID string) ([]pds.ModelsServiceAccount, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	saModels, res, err := saClient.ApiTenantsIdServiceAccountsGet(sa.context, tenantID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdServiceAccountsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModels.GetData(), nil
}

// GetServiceAccount func
func (sa *ServiceAccount) GetServiceAccount(serviceAccountID string) (*pds.ControllersServiceAccountResponse, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	saModel, res, err := saClient.ApiServiceAccountsIdGet(sa.context, serviceAccountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModel, nil
}

// CreateServiceAccountToken func
func (sa *ServiceAccount) CreateServiceAccountToken(tenantID string, name string) (*pds.ModelsServiceAccount, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	createRequest := pds.ControllersCreateServiceAccountRequest{Name: &name}
	saModel, res, err := saClient.ApiTenantsIdServiceAccountsPost(sa.context, tenantID).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdTokenGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModel, nil
}

// GetServiceAccountToken func
func (sa *ServiceAccount) GetServiceAccountToken(serviceAccountID string) (*pds.ControllersServiceAccountTokenResponse, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	saModel, res, err := saClient.ApiServiceAccountsIdTokenGet(sa.context, serviceAccountID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdTokenGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModel, nil
}

// DeleteServiceAccount func
func (sa *ServiceAccount) DeleteServiceAccount(serviceAccountID string) (*status.Response, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	res, err := saClient.ApiServiceAccountsIdDelete(sa.context, serviceAccountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
