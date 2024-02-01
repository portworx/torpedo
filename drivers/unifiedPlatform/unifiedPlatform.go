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
			ApplicationV2: &platformApi.ApplicationsV2{
				ApiClientV2: platformApiClient,
			},
			BackupLocationV2: &platformApi.BackupLocationV2{
				ApiClientV2: platformApiClient,
			},
			CloudCredentialsV2: &platformApi.CloudCredentialsV2{
				ApiClientV2: platformApiClient,
			},
			InvitationV2: &platformApi.InvitationV2{
				ApiClientV2: platformApiClient,
			},
			ServiceAccountV2: &platformApi.ServiceAccountV2{
				ApiClientV2: platformApiClient,
			},
			TargetClusterRegistrationV2: &platformApi.TargetClusterRegistrationV2{
				ApiClientV2: platformApiClient,
			},
			WhoAmiV2: &platformApi.WhoAmiV2{
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
			BackupConfigV2: &pdsV2Api.BackupConfigV2{
				ApiClientV2: pdsApiClient,
			},
			BackupV2: &pdsV2Api.BackupV2{
				ApiClientV2: pdsApiClient,
			},
			RestoreV2: &pdsV2Api.RestoreV2{
				ApiClientV2: pdsApiClient,
			},
		},
	}
}
