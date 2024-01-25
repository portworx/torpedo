package unifiedControlPlane

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	platformApi "github.com/portworx/torpedo/drivers/unifiedControlPlane/platform/api"
)

type UnifiedControlPlaneComponents struct {
	Accountv2 *platformApi.Accountv2
}

func NewUnifiedControlPlane(apiClient *pdsv2.APIClient) *UnifiedControlPlaneComponents {
	return &UnifiedControlPlaneComponents{
		Accountv2: &platformApi.Accountv2{
			ApiClientv2: apiClient,
		},
	}
}
