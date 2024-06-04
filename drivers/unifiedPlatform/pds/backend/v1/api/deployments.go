package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"

	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	deploymentV1 "github.com/pure-px/platform-api-go-client/pds/v1/dataservicedeployment"
)

var (
	DeploymentRequest deploymentV1.V1DataServiceDeployment
)

func (ds *PDS_API_V1) GetDeployment(deploymentId string) (*automationModels.PDSDeploymentResponse, error) {
	dsResponse := automationModels.PDSDeploymentResponse{
		Get: automationModels.V1Deployment{},
	}
	ctx, dsClient, err := ds.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	dsModel, res, err := dsClient.DataServiceDeploymentServiceGetDataServiceDeployment(ctx, deploymentId).Execute()
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

	_, res, err := dsClient.DataServiceDeploymentServiceDeleteDataServiceDeployment(ctx, deploymentId).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `DataServiceDeploymentServiceDeleteDataServiceDeployment`: %v\n.Full HTTP response: %v", err, res)
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

	dsModel, res, err := dsClient.DataServiceDeploymentServiceListDataServiceDeployments(ctx).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DataServiceDeploymentServiceCreateDataServiceDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(dsModel.DataServiceDeployments, &dsResponse.List)
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
	depCreateRequest := deploymentV1.ApiDataServiceDeploymentServiceCreateDataServiceDeploymentRequest{}

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
	fmt.Println("Storage Template Id: ", *DeploymentRequest.Config.DataServiceDeploymentTopologies[0].StorageOptions.Id)
	fmt.Println("App Template Id: ", *DeploymentRequest.Config.DataServiceDeploymentTopologies[0].ServiceConfigurations.Id)
	fmt.Println("Resource Template Id: ", *DeploymentRequest.Config.DataServiceDeploymentTopologies[0].ResourceSettings.Id)

	DeploymentRequestBody := deploymentV1.DataServiceDeploymentServiceCreateDataServiceDeploymentBody{
		ProjectId:             &createDeploymentRequest.Create.ProjectID,
		DataServiceDeployment: &DeploymentRequest,
	}

	depCreateRequest = dsClient.DataServiceDeploymentServiceCreateDataServiceDeployment(ctx, createDeploymentRequest.Create.NamespaceID).DataServiceDeploymentServiceCreateDataServiceDeploymentBody(DeploymentRequestBody)
	dsModel, res, err := dsClient.DataServiceDeploymentServiceCreateDataServiceDeploymentExecute(depCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DataServiceDeploymentServiceCreateDataServiceDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	log.Debugf("deployment Name [%s]", *dsModel.Status.CustomResourceName)

	err = utilities.CopyStruct(dsModel, &dsResponse.Create)
	if err != nil {
		return nil, fmt.Errorf("Error while copying create deployment response: %v\n", err)
	}
	return &dsResponse, err
}

func (ds *PDS_API_V1) GetDeploymentCredentials(deploymentId string) (string, error) {
	ctx, dsClient, err := ds.getDeploymentClient()
	if err != nil {
		return "", fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	dsModel, res, err := dsClient.DataServiceDeploymentServiceGetDataServiceDeploymentCredentials(ctx, deploymentId).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return "", fmt.Errorf("Error when calling `DeploymentServiceGetDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	return *dsModel.Secret, nil
}
