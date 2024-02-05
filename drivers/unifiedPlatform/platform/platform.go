package platform

import (
	platformApi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api"
)

type Platform struct {
	AccountV2         *platformApi.AccountV2
	TargetClusterV2   *platformApi.TargetClusterV2
	IamRoleBindingsV2 *platformApi.IamRoleBindingsV2
	NamespaceV2       *platformApi.NamespaceV2
	ProjectV2         *platformApi.ProjectV2
	TenantV2          *platformApi.TenantV2
	WhoAmI            *platformApi.WhoAmI
}
