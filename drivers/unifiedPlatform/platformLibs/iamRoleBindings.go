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

// GetIamRoleBinding returns IAM role binding for given ActorId
func GetIamRoleBinding(actorId string) (*automationModels.IAMResponse, error) {
	iamImputs := &automationModels.IAMRequest{
		Get: automationModels.GetIAM{
			ActorId: actorId,
		},
	}

	iamModel, err := v2Components.Platform.GetIamRoleBindingByID(iamImputs)
	if err != nil {
		return nil, err
	}
	log.Infof("IAM Role - %v", iamModel)
	return iamModel, nil
}

// GrantIamAccess grants IAM roles to given ActorId
func GrantIamAccess(actorId string, nsRoleBindings map[string][]automationModels.V1RoleBinding, projectId string, tenantId string) error {

	iamInputs := &automationModels.IAMRequest{
		Grant: automationModels.GrantIAM{
			IamConfigActorId: actorId,
			IAMServiceGrantIAMBody: &automationModels.IAMServiceGrantIAMBody{
				Iam: &automationModels.V1AccessPolicy{},
			},
		},
	}

	if val, ok := nsRoleBindings[tenantAdminRole]; ok {
		iamInputs.Grant.IAMServiceGrantIAMBody.Iam.Tenant = val
		iamInputs.Grant.IAMServiceGrantIAMBody.TenantId = &tenantId
	}

	if val, ok := nsRoleBindings[projectWriterRole]; ok {
		iamInputs.Grant.IAMServiceGrantIAMBody.Iam.Project = val
		iamInputs.Grant.IAMServiceGrantIAMBody.ProjectId = &projectId
	}

	if val, ok := nsRoleBindings[projectAdminRole]; ok {
		iamInputs.Grant.IAMServiceGrantIAMBody.Iam.Project = val
		iamInputs.Grant.IAMServiceGrantIAMBody.ProjectId = &projectId
	}

	log.Infof("Granting IAM Roles - %v", iamInputs.Grant.IAMServiceGrantIAMBody)
	log.Infof("Granting IAM Roles Access - %v", iamInputs.Grant.IAMServiceGrantIAMBody.Iam.Project)

	iamModel, err := v2Components.Platform.GrantIAMRoles(iamInputs)
	if err != nil {
		return err
	}
	log.Infof("IAM Roles granted - [%s]", *iamModel.Grant.Message)
	return nil
}

// RevokeIamAccess revokes IAM roles to given ActorId
func RevokeIamAccess(actorId string, nsRoleBindings map[string][]automationModels.V1RoleBinding, projectId string, tenantId string) error {

	iamInputs := &automationModels.IAMRequest{
		Revoke: automationModels.RevokeIAM{
			IamConfigActorId: actorId,
			IAMServiceRevokeIAMBody: &automationModels.IAMServiceRevokeIAMBody{
				Iam: &automationModels.V1AccessPolicy{},
			},
		},
	}

	if val, ok := nsRoleBindings[tenantAdminRole]; ok {
		iamInputs.Revoke.IAMServiceRevokeIAMBody.Iam.Tenant = val
		iamInputs.Revoke.IAMServiceRevokeIAMBody.TenantId = &tenantId
	}

	if val, ok := nsRoleBindings[projectWriterRole]; ok {
		iamInputs.Revoke.IAMServiceRevokeIAMBody.Iam.Project = val
		iamInputs.Revoke.IAMServiceRevokeIAMBody.ProjectId = &projectId
	}

	if val, ok := nsRoleBindings[projectAdminRole]; ok {
		iamInputs.Revoke.IAMServiceRevokeIAMBody.Iam.Project = val
		iamInputs.Revoke.IAMServiceRevokeIAMBody.ProjectId = &projectId
	}

	iamModel, err := v2Components.Platform.RevokeAccessForIAM(iamInputs)
	if err != nil {
		return err
	}
	log.Infof("IAM Roles revoked - [%s]", *iamModel.Grant.Message)
	return nil
}
