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

// GetClient updates the header with bearer token and returns the new client
func (ds *PDSV2_API) GetDeploymentConfigClient() (context.Context, *pdsv2.DeploymentConfigUpdateServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.ApiClientV2.DeploymentConfigUpdateServiceAPI

	return ctx, client, nil
}

func (ds *PDSV2_API) UpdateDeployment(updateDeploymentRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	var updateRequest pdsv2.ApiDeploymentConfigUpdateServiceCreateDeploymentConfigUpdateRequest
	dsResponse := apiStructs.WorkFlowResponse{}

	_, dsClient, err := ds.GetDeploymentConfigClient()
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
