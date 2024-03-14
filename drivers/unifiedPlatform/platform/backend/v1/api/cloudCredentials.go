package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	cloudCredentialv1 "github.com/pure-px/platform-api-go-client/platform/v1/cloudcredential"
	status "net/http"
)

// ListCloudCredentials return list of cloud credentials
func (cloudCred *PLATFORM_API_V1) ListCloudCredentials(request *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	cloudCredsResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceListCloudCredentials(ctx).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceListcloudCredations`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	err = copier.Copy(&cloudCredsResponse, cloudCredModel.CloudCredentials)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return cloudCredsResponse, nil
}

// GetCloudCredentials gets cloud credentials by ts id
func (cloudCred *PLATFORM_API_V1) GetCloudCredentials(getReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	var getCloudCredReq cloudCredentialv1.ApiCloudCredentialServiceGetCloudCredentialRequest
	err = copier.Copy(&getCloudCredReq, getReq)
	if err != nil {
		return nil, err
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceGetCloudCredentialExecute(getCloudCredReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceGetCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	err = copier.Copy(&cloudCredsResponse, cloudCredModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredModel)
	return &cloudCredsResponse, nil
}

// CreateCloudCredentials return newly created cloud credentials
func (cloudCred *PLATFORM_API_V1) CreateCloudCredentials(createRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	var createCloudCredRequest cloudCredentialv1.ApiCloudCredentialServiceCreateCloudCredentialRequest
	err = copier.Copy(&createCloudCredRequest, createRequest)
	if err != nil {
		return nil, err
	}
	cloudCredModel, _, err := cloudCredsClient.CloudCredentialServiceCreateCloudCredentialExecute(createCloudCredRequest)
	if err != nil {
		return nil, fmt.Errorf("error when called `CloudCredentialServiceCreateCloudCredential` to create cloud credential - %v", err)
	}
	err = copier.Copy(&cloudCredsResponse, cloudCredModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// UpdateCloudCredentials return updated created cloud credentials
func (cloudCred *PLATFORM_API_V1) UpdateCloudCredentials(updateReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	cloudCredsResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateAppReq cloudCredentialv1.ApiCloudCredentialServiceUpdateCloudCredentialRequest
	err = copier.Copy(&updateAppReq, updateReq)
	if err != nil {
		return nil, err
	}
	cloudCredationModel, res, err := cloudCredsClient.CloudCredentialServiceUpdateCloudCredentialExecute(updateAppReq)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceUpdatecloudCredation`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&cloudCredsResponse, cloudCredationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// DeleteCloudCredential delete cloud cred model.
func (cloudCred *PLATFORM_API_V1) DeleteCloudCredential(cloudCredId *WorkFlowRequest) error {
	ctx, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := cloudCredsClient.CloudCredentialServiceDeleteCloudCredential(ctx, cloudCredId.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `CloudCredentialServiceDeleteCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}
