package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

type DeploymentConfigurationUpdateV2 struct {
	ApiClientV2 *pdsv2.APIClient
}

// ListDeploymentConfigurationUpdates return deployments config update models for a given project.
func (depConfigUpdate *DeploymentConfigurationUpdateV2) ListDeploymentConfigurationUpdates() ([]pdsv2.V1DeploymentConfigUpdate, error) {
	dsClient := depConfigUpdate.ApiClientV2.DeploymentConfigUpdateServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModels, res, err := dsClient.DeploymentConfigUpdateServiceListDeploymentConfigUpdates(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceListDeploymentConfigUpdates`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModels.DeploymentConfigUpdates, nil
}

// CreateDeploymentConfigurationUpdate return newly created deployment model.
func (depConfigUpdate *DeploymentConfigurationUpdateV2) CreateDeploymentConfigurationUpdate(deploymentConfigUpdateConfigDeploymentMetaUid string) (*pdsv2.V1DeploymentConfigUpdate, error) {
	dsClient := depConfigUpdate.ApiClientV2.DeploymentConfigUpdateServiceAPI
	context, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentConfigUpdateServiceCreateDeploymentConfigUpdate(context, deploymentConfigUpdateConfigDeploymentMetaUid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceCreateDeploymentConfigUpdate`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

func (depConfigUpdate *DeploymentConfigurationUpdateV2) GetDeploymentConfigurationUpdate(deploymentID string) (*pdsv2.V1DeploymentConfigUpdate, error) {
	dsClient := depConfigUpdate.ApiClientV2.DeploymentConfigUpdateServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentConfigUpdateServiceGetDeploymentConfigUpdate(ctx, deploymentID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceGetDeploymentConfigUpdate`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

// RetryDeploymentConfigUpdate retries deployment status.
func (depConfigUpdate *DeploymentConfigurationUpdateV2) RetryDeploymentConfigUpdate(retryId string) (*pdsv2.V1DeploymentConfigUpdate, *status.Response, error) {
	dsClient := depConfigUpdate.ApiClientV2.DeploymentConfigUpdateServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentConfigUpdateServiceRetryDeploymentConfigUpdate(ctx, retryId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, nil, fmt.Errorf("Error when calling `DeploymentConfigUpdateServiceRetryDeploymentConfigUpdate`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, res, err
}
