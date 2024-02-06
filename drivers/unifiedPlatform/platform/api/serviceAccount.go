package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// ServiceAccountV2 struct
type ServiceAccountV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (sa *ServiceAccountV2) GetClient() (context.Context, *platformV2.ServiceAccountServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	sa.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	sa.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = sa.AccountID
	client := sa.ApiClientV2.ServiceAccountServiceAPI

	return ctx, client, nil
}

// ListAllServiceAccounts List all Service Accounts
func (sa *ServiceAccountV2) ListAllServiceAccounts() ([]platformV2.V1ServiceAccount, error) {
	ctx, client, err := sa.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saList, res, err := client.ServiceAccountServiceListServiceAccount(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ServiceAccountServiceListServiceAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return saList.ServiceAccounts, nil
}

// GetServiceAccount return service account model.
func (sa *ServiceAccountV2) GetServiceAccount(saID string) (*platformV2.V1ServiceAccount, error) {
	ctx, client, err := sa.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saModel, res, err := client.ServiceAccountServiceGetServiceAccount(ctx, saID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ServiceAccountServiceGetServiceAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return saModel, nil
}

// CreateServiceAccount return new service account model.
func (sa *ServiceAccountV2) CreateServiceAccount(tenantId string) (*platformV2.V1ServiceAccount, error) {
	ctx, client, err := sa.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saModel, res, err := client.ServiceAccountServiceCreateServiceAccount(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ServiceAccountServiceCreateServiceAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return saModel, nil
}

// DeleteServiceAccount delete service account and return status.
func (sa *ServiceAccountV2) DeleteServiceAccount(saId string) (*status.Response, error) {
	ctx, client, err := sa.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := client.ServiceAccountServiceDeleteServiceAccount(ctx, saId).Execute()
	if err != nil {
		return nil, fmt.Errorf("Error when calling `ServiceAccountServiceDeleteServiceAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}

// RegenerateServiceAccountSecret serviceAccountSecret
func (sa *ServiceAccountV2) RegenerateServiceAccountSecret(saId string) (*platformV2.V1ServiceAccount, error) {
	ctx, client, err := sa.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saModel, res, err := client.ServiceAccountServiceRegenerateServiceAccountSecret(ctx, saId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ServiceAccountServiceRegenerateServiceAccountSecret`: %v\n.Full HTTP response: %v", err, res)
	}
	return saModel, nil
}

// UpdateServiceAccount update existing serviceAccount
func (sa *ServiceAccountV2) UpdateServiceAccount(saId string) (*platformV2.V1ServiceAccount, error) {
	ctx, client, err := sa.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saModel, res, err := client.ServiceAccountServiceUpdateServiceAccount(ctx, saId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ServiceAccountServiceUpdateServiceAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	return saModel, nil
}
