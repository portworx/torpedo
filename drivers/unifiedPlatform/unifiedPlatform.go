package unifiedPlatform

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	platformApi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api"
)

type UnifiedPlatformComponents struct {
	Platform *platform.Platform
}

func NewUnifiedPlatformComponents(apiClient *pdsv2.APIClient) *UnifiedPlatformComponents {
	return &UnifiedPlatformComponents{
		Platform: &platform.Platform{
			Accountv2: &platformApi.Accountv2{
				ApiClientv2: apiClient,
			},
		},
	}
}
