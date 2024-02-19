package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetCloudCredentialClient updates the header with bearer token and return cloudCreds the new client
func (cloudCred *PLATFORM_API_V1) GetCloudCredentialClient() (context.Context, *platformv1.CloudCredentialServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	cloudCred.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	cloudCred.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = cloudCred.AccountID
	client := cloudCred.ApiClientV1.CloudCredentialServiceAPI

	return ctx, client, nil
}

// ListCloudCredentials return list of cloud credentials
func (cloudCred *PLATFORM_API_V1) ListCloudCredentials() ([]WorkFlowResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.GetCloudCredentialClient()
	cloudCredsResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceListCloudCredentials(ctx).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceListcloudCredations`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	copier.Copy(&cloudCredsResponse, cloudCredModel.CloudCredentials)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return cloudCredsResponse, nil
}

// GetCloudCredentials gets cloud credentials by ts id
func (cloudCred *PLATFORM_API_V1) GetCloudCredentials(cloudCredId string) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.GetCloudCredentialClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	var getRequest platformv1.ApiCloudCredentialServiceGetCloudCredentialRequest
	getRequest = getRequest.ApiService.CloudCredentialServiceGetCloudCredential(ctx, cloudCredId)
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceGetCloudCredential(ctx, cloudCredId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceGetCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	copier.Copy(&cloudCredsResponse, cloudCredModel)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredModel)
	return &cloudCredsResponse, nil
}

// CreateCloudCredentials return newly created cloud credentials
func (cloudCred *PLATFORM_API_V1) CreateCloudCredentials(createRequest platformv1.ApiCloudCredentialServiceCreateCloudCredentialRequest, tenantId string) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.GetCloudCredentialClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	createRequest = createRequest.ApiService.CloudCredentialServiceCreateCloudCredential(ctx, tenantId)
	cloudCredModel, _, err := cloudCredsClient.CloudCredentialServiceCreateCloudCredentialExecute(createRequest)
	if err != nil {
		return nil, fmt.Errorf("error when called `CloudCredentialServiceCreateCloudCredential` to create cloud credential - %v", err)
	}
	copier.Copy(&cloudCredsResponse, cloudCredModel)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// UpdateCloudCredentials return updated created cloud credentials
func (cloudCred *PLATFORM_API_V1) UpdateCloudCredentials(updateReq platformv1.ApiCloudCredentialServiceUpdateCloudCredentialRequest, cloudCredentialId string) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.GetCloudCredentialClient()
	cloudCredsResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	updateReq = updateReq.ApiService.CloudCredentialServiceUpdateCloudCredential(ctx, cloudCredentialId)
	cloudCredationModel, res, err := cloudCredsClient.CloudCredentialServiceUpdateCloudCredentialExecute(updateReq)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceUpdatecloudCredation`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&cloudCredsResponse, cloudCredationModel)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// DeleteCloudCredential delete cloud cred model.
func (cloudCred *PLATFORM_API_V1) DeleteCloudCredential(cloudCredId string) error {
	ctx, cloudCredsClient, err := cloudCred.GetCloudCredentialClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := cloudCredsClient.CloudCredentialServiceDeleteCloudCredential(ctx, cloudCredId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `CloudCredentialServiceDeleteCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}
