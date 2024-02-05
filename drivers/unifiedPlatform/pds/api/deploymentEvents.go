package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// DeploymentEventsV2 struct
type DeploymentEventsV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (de *DeploymentEventsV2) GetClient() (context.Context, *pdsv2.DeploymentEventServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	de.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	de.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = de.AccountID
	client := de.ApiClientV2.DeploymentEventServiceAPI

	return ctx, client, nil
}

// ListDeploymentEvents return deployments models for a given project.
func (de *DeploymentEventsV2) ListDeploymentEvents(deploymentId string) ([]pdsv2.V1DeploymentEvent, error) {
	ctx, deClient, err := de.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModels, res, err := deClient.DeploymentEventServiceListDeploymentEvents(ctx, deploymentId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentEventServiceListDeploymentEvents`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModels.DeploymentEvents, nil
}
