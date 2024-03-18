package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	"time"
)

const (
	targetClusterHealthOK = "CONNECTED"
)

// GetManifest Get the manifest for the account and tenant-id that can be used to install the platform agent
func GetManifest(tenantId string, clusterName string) (*automationModels.WorkFlowResponse, error) {

	response := &automationModels.WorkFlowResponse{
		TargetCluster: automationModels.PlatformTargetClusterOutput{
			Manifest: automationModels.PlatformManifestOutput{}}}

	manifestInputs := automationModels.PlatformTargetCluster{
		GetManifest: automationModels.PlatformGetTargetClusterManifest{
			ClusterName: clusterName,
			TenantId:    tenantId,
		},
	}

	// TODO: Proxy and Registry configs need to be added to this call

	if clusterName == "" {
		clusterName = fmt.Sprintf("Cluster_%v", time.Now().Unix())
	}

	log.Infof("cluster name [%s]", manifestInputs.GetManifest.ClusterName)

	// Get Manifest from API
	manifest, err := v2Components.Platform.GetTargetClusterRegistrationManifest(&manifestInputs)
	if err != nil {
		return response, err
	}

	response.TargetCluster.Manifest.Manifest = manifest.TargetCluster.Manifest.Manifest

	return response, nil
}

func ListTargetClusters(tenantId string) ([]automationModels.WorkFlowResponse, error) {
	wfRequest := automationModels.PlatformTargetCluster{
		ListTargetClusters: automationModels.PlatformListTargetCluster{
			TenantId: tenantId,
		},
	}

	tcList, err := v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return tcList, err
	}
	return tcList, nil
}
