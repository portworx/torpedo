package unifiedPlatform

import (
	pdsV2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/pds"
	pdsV2Api "github.com/portworx/torpedo/drivers/unifiedPlatform/pds/api"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	platformApi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

type UnifiedPlatformComponents struct {
	Platform *platform.Platform
	ApiV2Pds *pds.Pds
}

func NewUnifiedPlatformComponents(platformApiClient *platformV2.APIClient, pdsApiClient *pdsV2.APIClient) *UnifiedPlatformComponents {
	return &UnifiedPlatformComponents{
		Platform: &platform.Platform{
			AccountV2: &platformApi.AccountV2{
				ApiClientV2: platformApiClient,
			},
			TargetClusterV2: &platformApi.TargetClusterV2{
				ApiClientV2: platformApiClient,
			},
			IamRoleBindingsV2: &platformApi.IamRoleBindingsV2{
				ApiClientV2: platformApiClient,
			},
			NamespaceV2: &platformApi.NamespaceV2{
				ApiClientV2: platformApiClient,
			},
			ProjectV2: &platformApi.ProjectV2{
				ApiClientV2: platformApiClient,
			},
			TenantV2: &platformApi.TenantV2{
				ApiClientV2: platformApiClient,
			},
			TargetClusterManifestV2: &platformApi.TargetClusterManifestV2{
				ApiClientV2: platformApiClient,
			},
		},
		ApiV2Pds: &pds.Pds{
			DataServiceV2: &pdsV2Api.DataServiceV2{
				ApiClientV2: pdsApiClient,
			},
			DataServiceVersionsV2: &pdsV2Api.DataServiceVersionsV2{
				ApiClientV2: pdsApiClient,
			},
			DeploymentV2: &pdsV2Api.DeploymentV2{
				ApiClientV2: pdsApiClient,
			},
			DeploymentConfigurationUpdateV2: &pdsV2Api.DeploymentConfigurationUpdateV2{
				ApiClientV2: pdsApiClient,
			},
			DeploymentEventsV2: &pdsV2Api.DeploymentEventsV2{
				ApiClientV2: pdsApiClient,
			},
			DeploymentManifestV2: &pdsV2Api.DeploymentManifestV2{
				ApiClientV2: pdsApiClient,
			},
		},
	}
}
