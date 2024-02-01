package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// TargetClusterV2 struct
type TargetClusterV2 struct {
	ApiClientV2 *platformV2.APIClient
}

// ListTargetClusters return deployment targets models.
func (dt *TargetClusterV2) ListTargetClusters() ([]platformV2.V1TargetCluster, error) {
	dtClient := dt.ApiClientV2.TargetClusterServiceAPI
	ctx, err := GetContext()
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
	dtClient := dt.ApiClientV2.TargetClusterServiceAPI
	log.Infof("Get cluster details having uuid - %v", targetID)
	ctx, err := GetContext()
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
	dtClient := dt.ApiClientV2.TargetClusterServiceAPI
	log.Infof("Get cluster details having uuid - %v", targetClusterMetaUid)
	ctx, err := GetContext()
	dtModel, res, err := dtClient.TargetClusterServiceUpdateTargetCluster(ctx, targetClusterMetaUid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceUpdateTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return dtModel, nil
}

// DeleteTarget delete the deployment target and return status.
func (dt *TargetClusterV2) DeleteTarget(targetID string) (*status.Response, error) {
	dtClient := dt.ApiClientV2.TargetClusterServiceAPI
	log.Infof("Get cluster details having uuid - %v", targetID)
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := dtClient.TargetClusterServiceDeleteTargetCluster(ctx, targetID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return res, fmt.Errorf("Error when calling `TargetClusterServiceDeleteTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}

// ListTargetClustersBelongsToTenant not available

// ListTargetClustersBelongsToProject not available
