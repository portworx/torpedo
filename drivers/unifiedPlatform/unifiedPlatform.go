package unifiedPlatform

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/pds"
	pdsV2Api "github.com/portworx/torpedo/drivers/unifiedPlatform/pds/api"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	platformApi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

type UnifiedPlatformComponents struct {
	Platform *platform.Platform
	ApiV2Pds *pds.Pds
}

func NewUnifiedPlatformComponents(platformApiClient *platformv2.APIClient, pdsApiClient *pdsv2.APIClient) *UnifiedPlatformComponents {
	return &UnifiedPlatformComponents{
		Platform: &platform.Platform{
			Accountv2: &platformApi.Accountv2{
				ApiClientv2: platformApiClient,
			},
		},
		ApiV2Pds: &pds.Pds{
			DeploymentV2: &pdsV2Api.DeploymentV2{
				ApiClientv2: pdsApiClient,
			},
		},
	}
}
