package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// CloudCredentialsV2 struct
type CloudCredentialsV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (cloudCreds *CloudCredentialsV2) GetClient() (context.Context, *platformV2.CloudCredentialServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	cloudCreds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	cloudCreds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = cloudCreds.AccountID
	client := cloudCreds.ApiClientV2.CloudCredentialServiceAPI

	return ctx, client, nil
}

// ListCloudCredentials return list of cloud credentials
func (cloudCreds *CloudCredentialsV2) ListCloudCredentials() ([]platformV2.V1CloudCredential, error) {
	ctx, cloudCredsClient, err := cloudCreds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceListCloudCredentials(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceListCloudCredentials`: %v\n.Full HTTP response: %v", err, res)
	}
	return cloudCredModel.CloudCredentials, nil
}

// GetCloudCredentials gets cloud credentials by ts id
func (cloudCreds *CloudCredentialsV2) GetCloudCredentials(cloudCredId string) (*platformV2.V1CloudCredential, error) {
	ctx, cloudCredsClient, err := cloudCreds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceGetCloudCredential(ctx, cloudCredId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceGetCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	return cloudCredModel, nil
}

// CreateCloudCredentials return newly created cloud credentials
func (cloudCreds *CloudCredentialsV2) CreateCloudCredentials(tenantId string) (*platformV2.V1CloudCredential, error) {
	ctx, cloudCredsClient, err := cloudCreds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceCreateCloudCredential(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceCreateCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	return cloudCredModel, nil
}

// UpdateCloudCredentials return updated created cloud credentials
func (cloudCreds *CloudCredentialsV2) UpdateCloudCredentials(cloudCredId string) (*platformV2.V1CloudCredential, error) {
	ctx, cloudCredsClient, err := cloudCreds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceUpdateCloudCredential(ctx, cloudCredId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceUpdateCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	return cloudCredModel, nil
}

// DeleteCloudCredential delete cloud cred model.
func (cloudCreds *CloudCredentialsV2) DeleteCloudCredential(cloudCredId string) (*status.Response, error) {
	ctx, cloudCredsClient, err := cloudCreds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := cloudCredsClient.CloudCredentialServiceDeleteCloudCredential(ctx, cloudCredId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return res, fmt.Errorf("Error when calling `CloudCredentialServiceDeleteCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
