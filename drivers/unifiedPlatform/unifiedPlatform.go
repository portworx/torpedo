package unifiedPlatform

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/pds"
	pdsV2Api "github.com/portworx/torpedo/drivers/unifiedPlatform/pds/api"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	platformApi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api"
)

type UnifiedPlatformComponents struct {
	Platform *platform.Platform
	ApiV2Pds *pds.Pds
}

func NewUnifiedPlatformComponents(apiClient *pdsv2.APIClient) *UnifiedPlatformComponents {
	return &UnifiedPlatformComponents{
		Platform: &platform.Platform{
			Accountv2: &platformApi.Accountv2{
				ApiClientv2: apiClient,
			},
			BackupLocationV2: &platformApi.BackupLocationV2{
				ApiClientv2: apiClient,
			},
			TargetClusterV2: &platformApi.TargetClusterV2{
				ApiClientv2: apiClient,
			},
			IamRoleBindingsV2: &platformApi.IamRoleBindingsV2{
				ApiClientv2: apiClient,
			},
			NamespaceV2: &platformApi.NamespaceV2{
				ApiClientv2: apiClient,
			},
			ProjectV2: &platformApi.ProjectV2{
				ApiClientv2: apiClient,
			},
			TenantV2: &platformApi.TenantV2{
				ApiClientv2: apiClient,
			},
		},
		ApiV2Pds: &pds.Pds{
			BackupV2: &pdsV2Api.BackupV2{
				ApiClientv2: apiClient,
			},
			BackupConfigV2: &pdsV2Api.BackupConfigV2{
				ApiClientv2: apiClient,
			},
			DataServiceV2: &pdsV2Api.DataServiceV2{
				ApiClientv2: apiClient,
			},
			DataServiceVersionsV2: &pdsV2Api.DataServiceVersionsV2{
				ApiClientv2: apiClient,
			},
			DeploymentV2: &pdsV2Api.DeploymentV2{
				ApiClientv2: apiClient,
			},
			DeploymentConfigurationUpdateV2: &pdsV2Api.DeploymentConfigurationUpdateV2{
				ApiClientv2: apiClient,
			},
			DeploymentEventsV2: &pdsV2Api.DeploymentEventsV2{
				ApiClientv2: apiClient,
			},
			DeploymentManifestV2: &pdsV2Api.DeploymentManifestV2{
				ApiClientv2: apiClient,
			},
		},
	}
}
