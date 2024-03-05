package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// DeploymentV2 struct
type PDSV2_API struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

var (
	DeploymentRequestBody pdsv2.V1Deployment
)

// GetClient updates the header with bearer token and returns the new client
func (ds *PDSV2_API) GetDeploymentClient() (context.Context, *pdsv2.DeploymentServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.ApiClientV2.DeploymentServiceAPI

	return ctx, client, nil
}

func (ds *PDSV2_API) ListDeployment() (*apiStructs.WorkFlowResponse, error) {
	dsResponse := apiStructs.WorkFlowResponse{}

	return &dsResponse, nil
}

// CreateDeployment return newly created deployment model.
func (ds *PDSV2_API) CreateDeployment(createDeploymentRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	dsResponse := apiStructs.WorkFlowResponse{}
	depCreateRequest := pdsv2.ApiDeploymentServiceCreateDeploymentRequest{}

	_, dsClient, err := ds.GetDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	err = copier.Copy(&DeploymentRequestBody, createDeploymentRequest.Deployment.V1Deployment)
	if err != nil {
		return nil, fmt.Errorf("Error while copying the deployment request\n")
	}

	//Debug Print
	fmt.Println("DeploymentRequestBody Name ", *DeploymentRequestBody.Meta.Name)
	fmt.Println("Storage Template Id: ", *DeploymentRequestBody.Config.DeploymentTopologies[0].StorageTemplate.Id)
	fmt.Println("App Template Id: ", *DeploymentRequestBody.Config.DeploymentTopologies[0].ApplicationTemplate.Id)
	fmt.Println("Resource Template Id: ", *DeploymentRequestBody.Config.DeploymentTopologies[0].ResourceTemplate.Id)

	depCreateRequest = dsClient.DeploymentServiceCreateDeployment(context.Background(), createDeploymentRequest.Deployment.NamespaceID).V1Deployment(DeploymentRequestBody)

	dsModel, res, err := dsClient.DeploymentServiceCreateDeploymentExecute(depCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	copier.Copy(&dsResponse, dsModel)
	return &dsResponse, err
}
