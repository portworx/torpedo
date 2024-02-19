package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetClient updates the header with bearer token and returns the new client
func (tc *PLATFORM_API_V1) getTargetClusterClient() (context.Context, *platformv1.TargetClusterServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	tc.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	tc.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = tc.AccountID

	client := tc.ApiClientV1.TargetClusterServiceAPI
	return ctx, client, nil
}

// ListTargetClusters return deployment targets models.
func (tc *PLATFORM_API_V1) ListTargetClusters() ([]WorkFlowResponse, error) {
	tcResponse := []WorkFlowResponse{}
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModels, res, err := dtClient.TargetClusterServiceListTargetClusters(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceListTargetClusters`: %v\n.Full HTTP response: %v", err, res)
	}

	copier.Copy(&tcResponse, dtModels)

	return tcResponse, nil
}

// GetTarget return deployment target model.
func (tc *PLATFORM_API_V1) GetTarget(tcRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	tcResponse := WorkFlowResponse{}
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModel, res, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, tcRequest.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceGetTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&tcResponse, dtModel)
	return &tcResponse, nil
}

// PatchTargetCluster returns the updated the deployment target model
func (tc *PLATFORM_API_V1) PatchTargetCluster(tcRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	var patchRequest platformv1.ApiTargetClusterServiceUpdateTargetClusterRequest
	tcResponse := WorkFlowResponse{}
	_, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	copier.Copy(&patchRequest, tcRequest)

	dtModel, res, err := dtClient.TargetClusterServiceUpdateTargetClusterExecute(patchRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceUpdateTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}

	copier.Copy(&tcResponse, dtModel)
	return &tcResponse, nil
}

// DeleteTarget delete the deployment target and return status.
func (tc *PLATFORM_API_V1) DeleteTarget(tcRequest *WorkFlowRequest) error {
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := dtClient.TargetClusterServiceDeleteTargetCluster(ctx, tcRequest.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `TargetClusterServiceDeleteTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}

func (tc *PLATFORM_API_V1) GetClusterHealth(targetClusterId string) (*platformv1.PlatformTargetClusterv1Status, error) {
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	targetCluster, _, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, targetClusterId).Execute()
	log.Info("Get list of Accounts.")
	return targetCluster.Status, nil
}
