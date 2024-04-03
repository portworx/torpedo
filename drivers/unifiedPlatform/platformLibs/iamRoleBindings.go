package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	tenantAdminRole    = "tenant-admin"
	projectAdminRole   = "project-admin"
	namespaceAdminRole = "namespace-admin" // This is for future usage
)

// CreatePlatformServiceAccountIamRoles creates IAM roles for given Namespace role binding and ActorId
func CreatePlatformServiceAccountIamRoles(iamName, actorId string, nsRoleBindings map[string][]automationModels.V1RoleBinding) (*automationModels.PDSIAMResponse, error) {

	iamInputs := &automationModels.PDSIAMRequest{
		Create: automationModels.CreateIAM{
			V1IAM: automationModels.V1IAM{
				Meta: automationModels.V1Meta{},
				Config: automationModels.V1Config3{
					AccessPolicy: &automationModels.V1AccessPolicy{},
				},
			},
		},
	}

	iamInputs.Create.V1IAM.Meta.Name = &iamName
	iamInputs.Create.V1IAM.Config.ActorId = &actorId

	if val, ok := nsRoleBindings[tenantAdminRole]; ok {
		iamInputs.Create.V1IAM.Config.AccessPolicy.Tenant = val
	}

	if val, ok := nsRoleBindings[projectAdminRole]; ok {
		iamInputs.Create.V1IAM.Config.AccessPolicy.Project = val
	}

	if val, ok := nsRoleBindings[namespaceAdminRole]; ok {
		iamInputs.Create.V1IAM.Config.AccessPolicy.Namespace = val
	}

	iamModel, err := v2Components.Platform.CreateIamRoleBinding(iamInputs)
	if err != nil {
		return nil, err
	}
	log.InfoD("IAM Roles created - %v", iamModel)
	return iamModel, nil
}
