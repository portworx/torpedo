package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

var (
	iamInputs *apiStructs.WorkFlowRequest
)

// CreatePlatformServiceAccountIamRoles creates IAM roles for given Namespace role binding and ActorId
func CreatePlatformServiceAccountIamRoles(iamName, actorId string, nsRoleBindings []apiStructs.V1RoleBinding) (*apiStructs.WorkFlowResponse, error) {
	iamInputs.CreateIAM.V1IAM.Meta.Name = &iamName
	iamInputs.CreateIAM.V1IAM.Config.ActorId = &actorId
	iamInputs.CreateIAM.V1IAM.Config.AccessPolicy.Namespace = nsRoleBindings
	iamModel, err := v2Components.Platform.CreateIamRoleBinding(iamInputs)
	if err != nil {
		return nil, err
	}
	return iamModel, nil
}
