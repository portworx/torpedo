package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	deploymentsConfigUpdateV1 "github.com/pure-px/platform-api-go-client/pds/v1/deploymentconfigupdate"
	status "net/http"
)

func (ds *PDS_API_V2) UpdateDeployment(updateDeploymentRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	var updateRequest deploymentsConfigUpdateV1.ApiDeploymentConfigUpdateServiceCreateDeploymentConfigUpdateRequest
	dsResponse := apiStructs.WorkFlowResponse{}

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
