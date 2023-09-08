package api

import (
	"fmt"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/torpedo/drivers/pds/pdsutils"
)

// ServiceIdentity struct
type ServiceIdentity struct {
	apiClient *pds.APIClient
}

// ListServiceIdentities return service identities models for a project.
func (si *ServiceIdentity) ListServiceIdentities(tenantID string) ([]pds.ModelsServiceIdentity, error) {
	siClient := si.apiClient.ServiceIdentityApi
	ctx, err := pdsutils.GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	siModels, res, err := siClient.ApiAccountsIdServiceIdentityGet(ctx, tenantID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiAccountsIdServiceIdentityGet`: %v\n.Full HTTP response: %v", err, res)
	}
	return siModels.GetData(), nil
}

// CreateServiceIdentity returns newly create service identity object
func (si *ServiceIdentity) CreateServiceIdentity(tenantID string, name string) (*pds.ModelsServiceIdentityWithToken, error) {
	siClient := si.apiClient.ServiceIdentityApi
	createRequest := pds.RequestsServiceIdentityRequest{Name: name}
	ctx, err := pdsutils.GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	siModels, res, err := siClient.ApiAccountsIdServiceIdentityPost(ctx, tenantID).Body(createRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiAccountsIdServiceIdentityPost`: %v\n.Full HTTP response: %v", err, res)
	}
	return siModels, nil
}

// GetServiceIdentityByID return service identity model.
func (si *ServiceIdentity) GetServiceIdentityByID(serviceIdentityID string) (*pds.ModelsServiceIdentity, error) {
	siClient := si.apiClient.ServiceIdentityApi
	ctx, err := pdsutils.GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	siModel, res, err := siClient.ApiServiceIdentityIdGet(ctx, serviceIdentityID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceIdentityIdGet`: %v\n.Full HTTP response: %v", err, res)
	}
	return siModel, nil
}

// DeleteServiceIdentity delete service identity and return status.
func (si *ServiceIdentity) DeleteServiceIdentity(serviceIdentityID string) (*status.Response, error) {
	siClient := si.apiClient.ServiceIdentityApi
	ctx, err := pdsutils.GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	res, err := siClient.ApiServiceIdentityIdDelete(ctx, serviceIdentityID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceIdentityIdDelete`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}

// GenerateServiceIdentityToken generates new JWT token for a service Identity
func (si *ServiceIdentity) GenerateServiceIdentityToken(clientId string) (*pds.ControllersGenerateTokenResponse, error) {
	siClient := si.apiClient.ServiceIdentityApi
	createRequest := pds.ControllersGenerateTokenRequest{ClientId: &clientId}
	ctx, err := pdsutils.GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saModel, res, err := siClient.ServiceIdentityGenerateTokenPost(ctx).Body(createRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ServiceIdentityGenerateTokenPost`: %v\n.Full HTTP response: %v", err, res)
	}
	return saModel, nil
}

// GetServiceIdentityToken regenerates/ returns the created JWT token for the serviceIdentity
func (si *ServiceIdentity) GetServiceIdentityToken(siId string) (*pds.ModelsServiceIdentityWithToken, error) {
	siClient := si.apiClient.ServiceIdentityApi
	ctx, err := pdsutils.GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saModel, res, err := siClient.ApiServiceIdentityIdRegenerateGet(ctx, siId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceIdentityIdRegenerateGet`: %v\n.Full HTTP response: %v", err, res)
	}
	return saModel, nil
}
