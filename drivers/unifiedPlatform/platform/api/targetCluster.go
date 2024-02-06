package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// TargetClusterV2 struct
type TargetClusterV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (dt *TargetClusterV2) GetClient() (context.Context, *platformV2.TargetClusterServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	dt.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	dt.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = dt.AccountID
	client := dt.ApiClientV2.TargetClusterServiceAPI
	return ctx, client, nil
}

// ListTargetClusters return deployment targets models.
func (dt *TargetClusterV2) ListTargetClusters() ([]platformV2.V1TargetCluster, error) {
	ctx, dtClient, err := dt.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModels, res, err := dtClient.TargetClusterServiceListTargetClusters(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceListTargetClusters`: %v\n.Full HTTP response: %v", err, res)
	}
	return dtModels.Clusters, nil
}

// GetTarget return deployment target model.
func (dt *TargetClusterV2) GetTarget(targetID string) (*platformV2.V1TargetCluster, error) {
	ctx, dtClient, err := dt.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModel, res, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, targetID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceGetTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return dtModel, nil
}

// PatchTargetCluster returns the updated the deployment target model
func (dt *TargetClusterV2) PatchTargetCluster(targetClusterMetaUid string) (*platformV2.V1TargetCluster, error) {
	ctx, dtClient, err := dt.GetClient()
	dtModel, res, err := dtClient.TargetClusterServiceUpdateTargetCluster(ctx, targetClusterMetaUid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceUpdateTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return dtModel, nil
}

// DeleteTarget delete the deployment target and return status.
func (dt *TargetClusterV2) DeleteTarget(targetID string) (*status.Response, error) {
	ctx, dtClient, err := dt.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := dtClient.TargetClusterServiceDeleteTargetCluster(ctx, targetID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return res, fmt.Errorf("Error when calling `TargetClusterServiceDeleteTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}

func (dt *TargetClusterV2) GetClusterHealth(targetClusterId string) (*platformV2.PlatformTargetClusterv1Status, error) {
	ctx, dtClient, err := dt.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	targetCluster, _, err := dtClient.TargetClusterServiceGetTargetCluster(ctx, targetClusterId).Execute()
	log.Info("Get list of Accounts.")
	return targetCluster.Status, nil
}

// ListTargetClustersBelongsToTenant not available

// ListTargetClustersBelongsToProject not available
