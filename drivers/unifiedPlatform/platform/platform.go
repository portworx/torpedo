package platform

import (
	platformApi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api"
)

type Platform struct {
	AccountV2                   *platformApi.AccountV2
	TargetClusterV2             *platformApi.TargetClusterV2
	IamRoleBindingsV2           *platformApi.IamRoleBindingsV2
	NamespaceV2                 *platformApi.NamespaceV2
	ProjectV2                   *platformApi.ProjectV2
	TenantV2                    *platformApi.TenantV2
	ApplicationV2               *platformApi.ApplicationsV2
	BackupLocationV2            *platformApi.BackupLocationV2
	CloudCredentialsV2          *platformApi.CloudCredentialsV2
	InvitationV2                *platformApi.InvitationV2
	ServiceAccountV2            *platformApi.ServiceAccountV2
	TargetClusterRegistrationV2 *platformApi.TargetClusterRegistrationV2
	WhoAmiV2                    *platformApi.WhoAmiV2
}
