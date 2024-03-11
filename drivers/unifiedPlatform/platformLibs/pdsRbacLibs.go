package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"strconv"
)

func CreateUser(saName, accId string) (*apiStructs.WorkFlowResponse, error) {
	log.InfoD("Creating Service Account...")
	tenantId, err := GetDefaultTenantId(accId)
	if err != nil {
		return nil, err
	}
	saAcc, err := CreateServiceAccountForRBAC(saName, tenantId)
	log.Infof("created service account with name %s", *saAcc.Meta.Name)
	return saAcc, nil
}

func AssignRoleBindingsToUser(saName, roleName, resourceId, accId string) (*apiStructs.WorkFlowResponse, error) {
	var (
		userModel apiStructs.PdsServiceAccount
		binding   apiStructs.V1RoleBinding
		roles     []apiStructs.V1RoleBinding
	)
	tenantId, err := GetDefaultTenantId(accId)
	user, err := GetServiceAccFromSaName(tenantId, saName)
	err = utilities.CopyStruct(user, userModel)
	actorID := *userModel.Config.ClientId
	clientSecret := *userModel.Config.ClientSecret
	binding.RoleName = &roleName
	binding.ResourceIds = append(binding.ResourceIds, resourceId)
	iamName := "iam-" + strconv.Itoa(rand.Int())
	roles = append(roles, binding)
	iamRoles, err := CreatePlatformServiceAccountIamRoles(iamName, actorID, roles)
	log.FailOnError(err, "error while creating iam roles")
	log.Infof("created iam role with name %s", *iamRoles.Meta.Name)
	tokenRes, err := GenerateServiceAccountAccessToken(tenantId, actorID, clientSecret)
	return tokenRes, nil
}
