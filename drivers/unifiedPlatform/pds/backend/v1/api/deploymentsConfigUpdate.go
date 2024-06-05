package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	deploymentsConfigUpdateV1 "github.com/pure-px/platform-api-go-client/pds/v1/dataservicedeploymentconfigupdate"
	status "net/http"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

var (
	UpdateDeploymentRequest deploymentsConfigUpdateV1.DataServiceDeploymentConfigUpdateOfTheDataServiceConfigDeploymentUpdateRequest
)

func (ds *PDS_API_V1) GetDeploymentConfig(updateDeploymentRequest *automationModels.PDSDeploymentRequest) (*automationModels.PDSDeploymentResponse, error) {
	dsResponse := automationModels.PDSDeploymentResponse{
		Update: automationModels.V1DeploymentUpdate{},
	}

	ctx, dsClient, err := ds.getDeploymentConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	dsModel, res, err := dsClient.DataServiceDeploymentConfigUpdateServiceGetDataServiceDeploymentConfigUpdate(ctx, updateDeploymentRequest.Update.DeploymentConfigId).Execute()
	log.Debugf("updated dsModel [%v]", dsModel)
	log.Debugf("response [%v]", res)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `UpdateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(dsModel, &dsResponse.Update)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying updateDeploymentResponse %v\n", err)
	}
	return &dsResponse, err

}

func (ds *PDS_API_V1) UpdateDeployment(updateDeploymentRequest *automationModels.PDSDeploymentRequest) (*automationModels.PDSDeploymentResponse, error) {
	dsResponse := automationModels.PDSDeploymentResponse{
		Update: automationModels.V1DeploymentUpdate{},
	}

	ctx, dsClient, err := ds.getDeploymentConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	err = utilities.CopyStruct(updateDeploymentRequest.Update.V1Deployment, &UpdateDeploymentRequest)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying updateDeploymentRequest %v\n", err)
	}

	dsModel, res, err := dsClient.DataServiceDeploymentConfigUpdateServiceCreateDataServiceDeploymentConfigUpdate(ctx, updateDeploymentRequest.Update.DeploymentID).DataServiceDeploymentConfigUpdateOfTheDataServiceConfigDeploymentUpdateRequest(UpdateDeploymentRequest).Execute()
	log.Debugf("updated dsModel [%v]", dsModel)
	log.Debugf("response [%v]", res)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `UpdateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(dsModel, &dsResponse.Update)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying updateDeploymentResponse %v\n", err)
	}

	return &dsResponse, err
}
