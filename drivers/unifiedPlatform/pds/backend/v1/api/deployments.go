package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"

	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	deploymentV1 "github.com/pure-px/platform-api-go-client/pds/v1/deployment"
)

var (
	DeploymentRequest deploymentV1.V1Deployment
)

func (ds *PDS_API_V1) GetDeployment(deploymentId string) (*automationModels.PDSDeploymentResponse, error) {
	dsResponse := automationModels.PDSDeploymentResponse{
		Get: automationModels.V1Deployment{},
	}
	ctx, dsClient, err := ds.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	dsModel, res, err := dsClient.DeploymentServiceGetDeployment(ctx, deploymentId).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceGetDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(dsModel, &dsResponse.Get)
	if err != nil {
		return nil, fmt.Errorf("Error while copying create deployment response: %v\n", err)
	}

	return &dsResponse, nil
}

func (ds *PDS_API_V1) DeleteDeployment(deploymentId string) error {
	ctx, dsClient, err := ds.getDeploymentClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	_, res, err := dsClient.DeploymentServiceDeleteDeployment(ctx, deploymentId).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil
}

func (ds *PDS_API_V1) ListDeployment(projectId string) (*automationModels.PDSDeploymentResponse, error) {
	dsResponse := automationModels.PDSDeploymentResponse{
		List: []automationModels.V1Deployment{},
	}
	ctx, dsClient, err := ds.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	dsModel, res, err := dsClient.DeploymentServiceListDeployments(ctx).ProjectId(projectId).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(dsModel.Deployments, &dsResponse.List)
	if err != nil {
		return nil, fmt.Errorf("Error while copying list deployment response: %v\n", err)
	}

	return &dsResponse, nil
}

// CreateDeployment return newly created deployment model.
func (ds *PDS_API_V1) CreateDeployment(createDeploymentRequest *automationModels.PDSDeploymentRequest) (*automationModels.PDSDeploymentResponse, error) {
	dsResponse := automationModels.PDSDeploymentResponse{
		Create: automationModels.V1Deployment{},
	}
	depCreateRequest := deploymentV1.ApiDeploymentServiceCreateDeploymentRequest{}

	ctx, dsClient, err := ds.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	err = copier.Copy(&DeploymentRequest, createDeploymentRequest.Create.V1Deployment)
	if err != nil {
		return nil, fmt.Errorf("Error while copying the deployment request\n")
	}

	//Debug Print
	fmt.Println("DeploymentRequestBody Name ", *DeploymentRequest.Meta.Name)
	fmt.Println("Storage Template Id: ", *DeploymentRequest.Config.DeploymentTopologies[0].StorageOptions.Id)
	fmt.Println("App Template Id: ", *DeploymentRequest.Config.DeploymentTopologies[0].ServiceConfigurations.Id)
	fmt.Println("Resource Template Id: ", *DeploymentRequest.Config.DeploymentTopologies[0].ResourceSettings.Id)

	DeploymentRequestBody := deploymentV1.DeploymentServiceCreateDeploymentBody{
		ProjectId:  &createDeploymentRequest.Create.ProjectID,
		Deployment: &DeploymentRequest,
	}

	depCreateRequest = dsClient.DeploymentServiceCreateDeployment(ctx, createDeploymentRequest.Create.NamespaceID).DeploymentServiceCreateDeploymentBody(DeploymentRequestBody)
	dsModel, res, err := dsClient.DeploymentServiceCreateDeploymentExecute(depCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	log.Debugf("deployment Name [%s]", *dsModel.Status.CustomResourceName)

	err = utilities.CopyStruct(dsModel, &dsResponse.Create)
	if err != nil {
		return nil, fmt.Errorf("Error while copying create deployment response: %v\n", err)
	}
	return &dsResponse, err
}
