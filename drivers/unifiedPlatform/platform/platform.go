package platform

import (
	platformApi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api"
)

type Platform struct {
	Accountv2         *platformApi.Accountv2
	BackupLocationV2  *platformApi.BackupLocationV2
	TargetClusterV2   *platformApi.TargetClusterV2
	IamRoleBindingsV2 *platformApi.IamRoleBindingsV2
	NamespaceV2       *platformApi.NamespaceV2
	ProjectV2         *platformApi.ProjectV2
	TenantV2          *platformApi.TenantV2
}
