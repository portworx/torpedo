package platformLibs

import (
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"strconv"
)

func RunWithRbac(accId, roleName, resourceId string) (string, error) {
	var (
		saAccModel     apiStructs.V1ServiceAccountResponse
		generatedToken apiStructs.V1AccessToken
		binding        apiStructs.V1RoleBinding
		roles          []apiStructs.V1RoleBinding
	)
	log.InfoD("Creating Service Account...")
	tenantId, err := GetDefaultTenantId(accId)
	saName := "sa-" + strconv.Itoa(rand.Int())
	saAcc, err := CreateServiceAccountForRBAC(saName, tenantId)
	log.Infof("created service account with name %s", *saAcc.Meta.Name)

	log.InfoD("Now Creating IAM Roles for the service account...")
	err = copier.Copy(&saAcc, saAccModel)
	if err != nil {
		return "", err
	}
	binding.RoleName = &roleName
	binding.ResourceIds = append(binding.ResourceIds, resourceId)
	iamName := "iam-" + strconv.Itoa(rand.Int())
	roles = append(roles, binding)
	actorID := *saAccModel.Config.ClientId
	iamRoles, err := CreatePlatformServiceAccountIamRoles(iamName, actorID, roles)
	log.FailOnError(err, "error while creating iam roles")
	log.Infof("created iam role with name %s", *iamRoles.Meta.Name)
	clientSecret := *saAccModel.Config.ClientSecret
	tokenRes, err := GenerateServiceAccountAccessToken(tenantId, actorID, clientSecret)
	copier.Copy(&tokenRes, generatedToken)
	jwt := generatedToken.Token
	return jwt, nil
}
