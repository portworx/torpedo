package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	"time"
)

const (
	targetClusterHealthOK = "CONNECTED"
)

// GetManifest Get the manifest for the account and tenant-id that can be used to install the platform agent
func GetManifest(tenantId string, clusterName string) (*apiStructs.WorkFlowResponse, error) {

	response := &apiStructs.WorkFlowResponse{
		TargetCluster: apiStructs.PlatformTargetClusterOutput{
			Manifest: apiStructs.PlatformManifestOutput{}}}

	manifestInputs := apiStructs.WorkFlowRequest{}

	// TODO: Proxy and Registry configs need to be added to this call

	if clusterName == "" {
		clusterName = fmt.Sprintf("Cluster_%v", time.Now().Unix())
	}

	manifestInputs.TargetCluster.GetManifest.ClusterName = clusterName
	manifestInputs.TargetCluster.GetManifest.TenantId = tenantId
	log.Infof("cluster name [%s]", manifestInputs.TargetCluster.GetManifest.ClusterName)

	// Get Manifest from API
	manifest, err := v2Components.Platform.GetTargetClusterRegistrationManifest(&manifestInputs)
	if err != nil {
		return response, err
	}

	response.TargetCluster.Manifest.Manifest = manifest.TargetCluster.Manifest.Manifest

	return response, nil
}

func ListTargetClusters(tenantId string) ([]apiStructs.WorkFlowResponse, error) {
	wfRequest := apiStructs.WorkFlowRequest{
		TargetCluster: apiStructs.PlatformTargetCluster{
			ListTargetClusters: apiStructs.PlatformListTargetCluster{},
		},
	}

	wfRequest.TargetCluster.ListTargetClusters.TenantId = tenantId
	tcList, err := v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return tcList, err
	}
	return tcList, nil
}
