package api

import (
	"fmt"
	status "net/http"

	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	deploymentV1 "github.com/pure-px/platform-api-go-client/pds/v1/deployment"
)

var (
	DeploymentRequest deploymentV1.V1Deployment
)

func (ds *PDS_API_V1) GetDeployment(deploymentId string) (*automationModels.WorkFlowResponse, error) {
	dsResponse := automationModels.WorkFlowResponse{}
	ctx, dsClient, err := ds.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	dsModel, res, err := dsClient.DeploymentServiceGetDeployment(ctx, deploymentId).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = copier.Copy(&dsResponse.PDSDeployment.V1Deployment, dsModel)
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

func (ds *PDS_API_V1) ListDeployment() (*automationModels.WorkFlowResponse, error) {
	dsResponse := automationModels.WorkFlowResponse{}

	return &dsResponse, nil
}

// CreateDeployment return newly created deployment model.
func (ds *PDS_API_V1) CreateDeployment(createDeploymentRequest *automationModels.PDSDeploymentRequest) (*automationModels.WorkFlowResponse, error) {
	dsResponse := automationModels.WorkFlowResponse{}
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
	fmt.Println("TargetClusterId: ", *DeploymentRequest.Config.References.TargetClusterId)

	DeploymentRequestBody := deploymentV1.DeploymentServiceCreateDeploymentBody{
		ProjectId:  &createDeploymentRequest.Create.ProjectID,
		Deployment: &DeploymentRequest,
	}

	//Debug Print
	fmt.Println("DeploymentRequest Name ", *DeploymentRequestBody.Deployment.Meta.Name)

	depCreateRequest = dsClient.DeploymentServiceCreateDeployment(ctx, createDeploymentRequest.Create.NamespaceID).DeploymentServiceCreateDeploymentBody(DeploymentRequestBody)

	dsModel, res, err := dsClient.DeploymentServiceCreateDeploymentExecute(depCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = copier.Copy(&dsResponse.PDSDeployment.V1Deployment, dsModel)
	if err != nil {
		return nil, fmt.Errorf("Error while copying create deployment response: %v\n", err)
	}
	return &dsResponse, err
}
