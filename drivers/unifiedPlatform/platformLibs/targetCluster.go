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
func GetManifest(tenantId string, clusterName string) (*automationModels.V1TargetClusterRegistrationManifest, error) {
	manifestInputs := automationModels.PlatformTargetClusterRequest{
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
		return nil, err
	}

	return &manifest.GetManifest, nil
}

func ListTargetClusters(tenantId string) (*automationModels.V1ListTargetClustersResponse, error) {
	wfRequest := automationModels.PlatformTargetClusterRequest{
		ListTargetClusters: automationModels.PlatformListTargetCluster{
			TenantId: tenantId,
		},
	}

	tcList, err := v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return &tcList.ListTargetClusters, err
	}

	totalRecords := *tcList.ListTargetClusters.Pagination.TotalRecords
	log.Infof("Total target clusters under [%s] are [%s]", tenantId, totalRecords)

	wfRequest = automationModels.PlatformTargetClusterRequest{
		ListTargetClusters: automationModels.PlatformListTargetCluster{
			TenantId:             tenantId,
			PaginationPageSize:   totalRecords,
			PaginationPageNumber: DEFAULT_PAGE_NUMBER,
			SortSortOrder:        DEFAULT_SORT_ORDER,
			SortSortBy:           DEFAULT_SORT_BY,
		},
	}

	tcList, err = v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return &tcList.ListTargetClusters, err
	}

	return &tcList.ListTargetClusters, nil
}

func GetTargetCluster(clusterId string) (*automationModels.V1TargetCluster, error) {
	wfRequest := automationModels.PlatformTargetClusterRequest{
		GetTargetCluster: automationModels.PlatformGetTargetCluster{
			Id: clusterId,
		},
	}
	tc, err := v2Components.Platform.GetTargetCluster(&wfRequest)
	if err != nil {
		return nil, err
	}
	return &tc.GetTargetCluster, nil
}
