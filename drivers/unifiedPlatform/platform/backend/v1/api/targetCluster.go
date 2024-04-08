package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	targetClusterv1 "github.com/pure-px/platform-api-go-client/platform/v1/targetcluster"
	status "net/http"
)

// ListTargetClusters return deployment targets models.
func (tc *PLATFORM_API_V1) ListTargetClusters(tcRequest *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error) {
	tcResponse := PlatformTargetClusterResponse{
		ListTargetClusters: V1ListTargetClustersResponse{},
	}
	tenantId := tcRequest.ListTargetClusters.TenantId
	_, dtClient, err := tc.getTargetClusterClient()

	var req targetClusterv1.ApiTargetClusterServiceListTargetClustersRequest
	req = req.TenantId(tenantId)
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	if tcRequest.ListTargetClusters.SortSortBy != "" {
		req = req.SortSortBy(tcRequest.ListTargetClusters.SortSortBy)
	}
	if tcRequest.ListTargetClusters.SortSortOrder != "" {
		req = req.SortSortOrder(tcRequest.ListTargetClusters.SortSortOrder)
	}
	if tcRequest.ListTargetClusters.PaginationPageNumber != "" {
		req = req.PaginationPageNumber(tcRequest.ListTargetClusters.PaginationPageNumber)
	}
	if tcRequest.ListTargetClusters.PaginationPageSize != "" {
		req = req.PaginationPageSize(tcRequest.ListTargetClusters.PaginationPageSize)
	}

	dtModels, res, err := dtClient.TargetClusterServiceListTargetClustersExecute(req)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceListTargetClusters`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(dtModels, &tcResponse.ListTargetClusters)
	if err != nil {
		return nil, err
	}

	return &tcResponse, nil
}

// GetTarget return deployment target model.
func (tc *PLATFORM_API_V1) GetTargetCluster(tcRequest *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error) {
	tcResponse := PlatformTargetClusterResponse{
		GetTargetCluster: V1TargetCluster{},
	}
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModel, res, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, tcRequest.GetTargetCluster.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceGetTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(dtModel, &tcResponse.GetTargetCluster)
	if err != nil {
		return nil, err
	}

	return &tcResponse, nil
}

// PatchTargetCluster returns the updated the deployment target model
//func (tc *PLATFORM_API_V1) PatchTargetCluster(tcRequest *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error) {
//	var patchRequest targetClusterv1.ApiTargetClusterServiceUpdateTargetClusterRequest
//	tcResponse := WorkFlowResponse{}
//	_, dtClient, err := tc.getTargetClusterClient()
//	if err != nil {
//		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
//	}
//
//	err = copier.Copy(&patchRequest, tcRequest)
//	if err != nil {
//		return nil, err
//	}
//
//	dtModel, res, err := dtClient.TargetClusterServiceUpdateTargetClusterExecute(patchRequest)
//	if err != nil && res.StatusCode != status.StatusOK {
//		return nil, fmt.Errorf("Error when calling `TargetClusterServiceUpdateTargetCluster`: %v\n.Full HTTP response: %v", err, res)
//	}
//
//	err = copier.Copy(&tcResponse, dtModel)
//	if err != nil {
//		return nil, err
//	}
//	return &tcResponse, nil
//}

// DeleteTargetCluster delete the deployment target and return status.
func (tc *PLATFORM_API_V1) DeleteTargetCluster(tcRequest *PlatformTargetClusterRequest) error {
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

func (tc *PLATFORM_API_V1) GetClusterHealth(tcRequest *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error) {
	ctx, dtClient, err := tc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	response := PlatformTargetClusterResponse{
		GetTargetClusterHealth: V1TargetCluster{},
	}

	targetCluster, res, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, tcRequest.GetTargetClusterHealth.Id).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceGetTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(targetCluster, &response.GetTargetClusterHealth)

	return &response, nil
}
