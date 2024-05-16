package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	tenantAdminRole   = "tenant-admin"
	projectAdminRole  = "project-admin"
	projectWriterRole = "project-writer"
)

// CreatePlatformServiceAccountIamRoles creates IAM roles for given Namespace role binding and ActorId
func CreatePlatformServiceAccountIamRoles(iamName, actorId string, nsRoleBindings map[string][]automationModels.V1RoleBinding) (*automationModels.IAMResponse, error) {

	iamInputs := &automationModels.IAMRequest{
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

	if val, ok := nsRoleBindings[projectWriterRole]; ok {
		iamInputs.Create.V1IAM.Config.AccessPolicy.Project = val
	}

	if val, ok := nsRoleBindings[projectAdminRole]; ok {
		iamInputs.Create.V1IAM.Config.AccessPolicy.Project = val
	}

	iamModel, err := v2Components.Platform.CreateIamRoleBinding(iamInputs)
	if err != nil {
		return nil, err
	}
	log.Infof("IAM Roles created - %v", iamModel)
	return iamModel, nil
}
