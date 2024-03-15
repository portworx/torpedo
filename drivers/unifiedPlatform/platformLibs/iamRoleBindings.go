package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

var (
	iamInputs *automationModels.WorkFlowRequest
)

// CreatePlatformServiceAccountIamRoles creates IAM roles for given Namespace role binding and ActorId
func CreatePlatformServiceAccountIamRoles(iamName, actorId string, nsRoleBindings []automationModels.V1RoleBinding) (*automationModels.WorkFlowResponse, error) {
	iamInputs.Iam.Create.V1IAM.Meta.Name = &iamName
	iamInputs.Iam.Create.V1IAM.Config.ActorId = &actorId
	iamInputs.Iam.Create.V1IAM.Config.AccessPolicy.Namespace = nsRoleBindings
	iamModel, err := v2Components.Platform.CreateIamRoleBinding(iamInputs)
	if err != nil {
		return nil, err
	}
	log.InfoD("IAM Roles created - %v", iamModel)
	return iamModel, nil
}
