package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// DeploymentTargetV2 struct
type DeploymentTargetV2 struct {
	apiClient *pdsv2.APIClient
}

// ListTargetClusters return deployment targets models.
func (dt *DeploymentTargetV2) ListTargetClusters() ([]pdsv2.V1TargetCluster, error) {
	dtClient := dt.apiClient.TargetClusterServiceApi
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
func (dt *DeploymentTargetV2) GetTarget(targetID string) (*pdsv2.V1TargetCluster, error) {
	dtClient := dt.apiClient.TargetClusterServiceApi
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

// PatchDeploymentTarget returns the updated the deployment target model
func (dt *DeploymentTargetV2) PatchDeploymentTarget(targetClusterMetaUid string) (*pdsv2.V1TargetCluster, error) {
	dtClient := dt.apiClient.TargetClusterServiceApi
	log.Infof("Get cluster details having uuid - %v", targetClusterMetaUid)
	ctx, err := GetContext()
	dtModel, res, err := dtClient.TargetClusterServiceUpdateTargetCluster(ctx, targetClusterMetaUid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceUpdateTargetCluster`: %v\n.Full HTTP response: %v", err, res)
	}
	return dtModel, nil
}

// DeleteTarget delete the deployment target and return status.
func (dt *DeploymentTargetV2) DeleteTarget(targetID string) (*status.Response, error) {
	dtClient := dt.apiClient.TargetClusterServiceApi
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

// ListDeploymentTargetsBelongsToTenant not available

// ListDeploymentTargetsBelongsToProject not available
