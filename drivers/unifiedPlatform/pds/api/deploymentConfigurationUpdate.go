package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

type DeploymentConfigurationUpdateV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (depConfigUpdate *DeploymentConfigurationUpdateV2) GetClient() (context.Context, *pdsv2.DeploymentConfigUpdateServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	depConfigUpdate.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	depConfigUpdate.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = depConfigUpdate.AccountID
	client := depConfigUpdate.ApiClientV2.DeploymentConfigUpdateServiceAPI

	return ctx, client, nil
}

// ListDeploymentConfigurationUpdates return deployments config update models for a given project.
func (depConfigUpdate *DeploymentConfigurationUpdateV2) ListDeploymentConfigurationUpdates() ([]pdsv2.V1DeploymentConfigUpdate, error) {
	ctx, dcpClient, err := depConfigUpdate.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModels, res, err := dcpClient.DeploymentConfigUpdateServiceListDeploymentConfigUpdates(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceListDeploymentConfigUpdates`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModels.DeploymentConfigUpdates, nil
}

// CreateDeploymentConfigurationUpdate return newly created deployment model.
func (depConfigUpdate *DeploymentConfigurationUpdateV2) CreateDeploymentConfigurationUpdate(deploymentConfigUpdateConfigDeploymentMetaUid string) (*pdsv2.V1DeploymentConfigUpdate, error) {
	context, dcpClient, err := depConfigUpdate.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dcpClient.DeploymentConfigUpdateServiceCreateDeploymentConfigUpdate(context, deploymentConfigUpdateConfigDeploymentMetaUid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceCreateDeploymentConfigUpdate`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

func (depConfigUpdate *DeploymentConfigurationUpdateV2) GetDeploymentConfigurationUpdate(deploymentID string) (*pdsv2.V1DeploymentConfigUpdate, error) {
	ctx, dcpClient, err := depConfigUpdate.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dcpClient.DeploymentConfigUpdateServiceGetDeploymentConfigUpdate(ctx, deploymentID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceGetDeploymentConfigUpdate`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

// RetryDeploymentConfigUpdate retries deployment status.
func (depConfigUpdate *DeploymentConfigurationUpdateV2) RetryDeploymentConfigUpdate(retryId string) (*pdsv2.V1DeploymentConfigUpdate, *status.Response, error) {
	ctx, dcpClient, err := depConfigUpdate.GetClient()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dcpClient.DeploymentConfigUpdateServiceRetryDeploymentConfigUpdate(ctx, retryId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceRetryDeploymentConfigUpdate`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, res, err
}
