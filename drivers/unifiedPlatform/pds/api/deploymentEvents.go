package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// DeploymentEventsV2 struct
type DeploymentEventsV2 struct {
	ApiClientv2 *pdsv2.APIClient
}

// ListDeploymentEvents return deployments models for a given project.
func (de *DeploymentEventsV2) ListDeploymentEvents(deploymentId string) ([]pdsv2.V1DeploymentEvent, error) {
	deClient := de.ApiClientv2.DeploymentEventServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModels, res, err := deClient.DeploymentEventServiceListDeploymentEvents(ctx, deploymentId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentEventServiceListDeploymentEvents`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModels.DeploymentEvents, nil
}
