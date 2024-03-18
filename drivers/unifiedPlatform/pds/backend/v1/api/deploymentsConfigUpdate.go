package api

import (
	"fmt"
	status "net/http"

	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"

	deploymentsConfigUpdateV1 "github.com/pure-px/platform-api-go-client/pds/v1/deploymentconfigupdate"
)

func (ds *PDS_API_V1) UpdateDeployment(updateDeploymentRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	var updateRequest deploymentsConfigUpdateV1.ApiDeploymentConfigUpdateServiceCreateDeploymentConfigUpdateRequest
	dsResponse := automationModels.WorkFlowResponse{}

	_, dsClient, err := ds.getDeploymentConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	copier.Copy(&updateRequest, updateDeploymentRequest)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying updateDeploymentRequest %v\n", err)
	}

	dsModel, res, err := dsClient.DeploymentConfigUpdateServiceCreateDeploymentConfigUpdateExecute(updateRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `UpdateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = copier.Copy(&dsResponse, dsModel)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying updateDeploymentResponse %v\n", err)
	}

	return &dsResponse, err
}
