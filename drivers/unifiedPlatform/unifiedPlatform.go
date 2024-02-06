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

func NewUnifiedPlatformComponents(platformApiClient *platformV2.APIClient, pdsApiClient *pdsV2.APIClient, AccountId string) *UnifiedPlatformComponents {
	return &UnifiedPlatformComponents{
		Platform: &platform.Platform{
			AccountV2: &platformApi.AccountV2{
				ApiClientV2: platformApiClient,
			},
			TargetClusterV2: &platformApi.TargetClusterV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			IamRoleBindingsV2: &platformApi.IamRoleBindingsV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			NamespaceV2: &platformApi.NamespaceV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			ProjectV2: &platformApi.ProjectV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			TenantV2: &platformApi.TenantV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			WhoAmI: &platformApi.WhoAmI{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			TargetClusterManifestV2: &platformApi.TargetClusterManifestV2{
				ApiClientV2: platformApiClient,
			},
			ApplicationV2: &platformApi.ApplicationsV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			BackupLocationV2: &platformApi.BackupLocationV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			CloudCredentialsV2: &platformApi.CloudCredentialsV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			InvitationV2: &platformApi.InvitationV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
			ServiceAccountV2: &platformApi.ServiceAccountV2{
				ApiClientV2: platformApiClient,
				AccountID:   AccountId,
			},
		},
		ApiV2Pds: &pds.Pds{
			DataServiceV2: &pdsV2Api.DataServiceV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			DataServiceVersionsV2: &pdsV2Api.DataServiceVersionsV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			DeploymentV2: &pdsV2Api.DeploymentV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			DeploymentConfigurationUpdateV2: &pdsV2Api.DeploymentConfigurationUpdateV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			DeploymentEventsV2: &pdsV2Api.DeploymentEventsV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			DeploymentManifestV2: &pdsV2Api.DeploymentManifestV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			BackupV2: &pdsV2Api.BackupV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			BackupConfigV2: &pdsV2Api.BackupConfigV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			DeploymentTopologyV2: &pdsV2Api.DeploymentTopologyV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			ImageV2: &pdsV2Api.ImageV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
			RestoreV2: &pdsV2Api.RestoreV2{
				ApiClientV2: pdsApiClient,
				AccountID:   AccountId,
			},
		},
	}
}
