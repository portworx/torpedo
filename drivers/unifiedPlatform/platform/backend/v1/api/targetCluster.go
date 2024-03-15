package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	targetClusterv1 "github.com/pure-px/platform-api-go-client/platform/v1/targetcluster"
	status "net/http"
)

// ListTargetClusters return deployment targets models.
func (tc *PLATFORM_API_V1) ListTargetClusters(tcRequest *PlatformTargetCluster) ([]WorkFlowResponse, error) {
	tcResponse := []WorkFlowResponse{}
	tenantId := tcRequest.ListTargetClusters.TenantId
	_, dtClient, err := tc.getTargetClusterClient()

	var req targetClusterv1.ApiTargetClusterServiceListTargetClustersRequest
	req = req.TenantId(tenantId)
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModels, res, err := dtClient.TargetClusterServiceListTargetClustersExecute(req)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceListTargetClusters`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&tcResponse, dtModels.Clusters)
	if err != nil {
		return nil, err
	}

	return tcResponse, nil
}

// GetTarget return deployment target model.
func (tc *PLATFORM_API_V1) GetTargetCluster(tcRequest *PlatformTargetCluster) (*WorkFlowResponse, error) {
	tcResponse := WorkFlowResponse{}
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModel, res, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, tcRequest.GetTargetCluster.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceGetTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&tcResponse, dtModel)
	if err != nil {
		return nil, err
	}
	return &tcResponse, nil
}

// PatchTargetCluster returns the updated the deployment target model
func (tc *PLATFORM_API_V1) PatchTargetCluster(tcRequest *PlatformTargetCluster) (*WorkFlowResponse, error) {
	var patchRequest targetClusterv1.ApiTargetClusterServiceUpdateTargetClusterRequest
	tcResponse := WorkFlowResponse{}
	_, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	err = copier.Copy(&patchRequest, tcRequest)
	if err != nil {
		return nil, err
	}

	dtModel, res, err := dtClient.TargetClusterServiceUpdateTargetClusterExecute(patchRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceUpdateTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}

	err = copier.Copy(&tcResponse, dtModel)
	if err != nil {
		return nil, err
	}
	return &tcResponse, nil
}

// DeleteTargetCluster delete the deployment target and return status.
func (tc *PLATFORM_API_V1) DeleteTargetCluster(tcRequest *PlatformTargetCluster) error {
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := dtClient.TargetClusterServiceDeleteTargetCluster(ctx, tcRequest.DeleteTargetCluster.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `TargetClusterServiceDeleteTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}

func (tc *PLATFORM_API_V1) GetClusterHealth(tcRequest *PlatformTargetCluster) (*targetClusterv1.PlatformTargetClusterv1Status, error) {
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	targetCluster, _, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, tcRequest.GetTargetClusterHealth.Id).Execute()
	log.Info("Get list of Accounts.")
	return targetCluster.Status, nil
}
