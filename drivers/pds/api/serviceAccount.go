package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type ServiceAccount struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (sa *ServiceAccount) ListServiceAccounts(tenantId string) ([]pds.ModelsServiceAccount, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	saModels, res, err := saClient.ApiTenantsIdServiceAccountsGet(sa.context, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdServiceAccountsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModels.GetData(), nil
}

func (sa *ServiceAccount) GetServiceAccount(serviceAccountId string) (*pds.ControllersServiceAccountResponse, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	saModel, res, err := saClient.ApiServiceAccountsIdGet(sa.context, serviceAccountId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModel, nil
}

func (sa *ServiceAccount) CreateServiceAccountToken(tenantId string, name string) (*pds.ModelsServiceAccount, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	createRequest := pds.ControllersCreateServiceAccountRequest{Name: &name}
	saModel, res, err := saClient.ApiTenantsIdServiceAccountsPost(sa.context, tenantId).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdTokenGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModel, nil
}

func (sa *ServiceAccount) GetServiceAccountToken(serviceAccountId string) (*pds.ControllersServiceAccountTokenResponse, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	saModel, res, err := saClient.ApiServiceAccountsIdTokenGet(sa.context, serviceAccountId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdTokenGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return saModel, nil
}

func (sa *ServiceAccount) DeleteServiceAccount(serviceAccountId string) (*status.Response, error) {
	saClient := sa.apiClient.ServiceAccountsApi
	res, err := saClient.ApiServiceAccountsIdDelete(sa.context, serviceAccountId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiServiceAccountsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
