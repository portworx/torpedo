package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/utilities"
	status "net/http"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"

	deploymentsConfigUpdateV1 "github.com/pure-px/platform-api-go-client/pds/v1/deploymentconfigupdate"
)

func (ds *PDS_API_V1) UpdateDeployment(updateDeploymentRequest *automationModels.PDSDeploymentRequest) (*automationModels.PDSDeploymentResponse, error) {
	var updateRequest deploymentsConfigUpdateV1.ApiDeploymentConfigUpdateServiceCreateDeploymentConfigUpdateRequest
	dsResponse := automationModels.PDSDeploymentResponse{
		Update: automationModels.V1Deployment{},
	}

	_, dsClient, err := ds.getDeploymentConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	err = utilities.CopyStruct(updateDeploymentRequest.Update, &updateRequest)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying updateDeploymentRequest %v\n", err)
	}

	dsModel, res, err := dsClient.DeploymentConfigUpdateServiceCreateDeploymentConfigUpdateExecute(updateRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `UpdateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(dsModel, &dsResponse.Update)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying updateDeploymentResponse %v\n", err)
	}

	return &dsResponse, err
}
