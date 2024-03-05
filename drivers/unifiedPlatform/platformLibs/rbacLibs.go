package platformLibs

import (
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"strconv"
)

func RunWithRbac(accName, roleName, resourceId string) (string, error) {
	var (
		saAccModel     apiStructs.ServiceAccountRequest
		generatedToken apiStructs.V1AccessToken
		binding        apiStructs.V1RoleBinding
		roles          []apiStructs.V1RoleBinding
	)
	log.InfoD("Creating Service Account...")
	accList, err := GetAccountListv1()
	if err != nil {
		return "", err
	}
	accountId := GetPlatformAccountID(accList, accName)
	tenantId, err := GetTenantId(accountId)
	saName := "sa-" + strconv.Itoa(rand.Int())
	saAcc, err := CreateServiceAccountForRBAC(saName, tenantId)
	log.Infof("created service account with name %s", *saAcc.Meta.Name)

	log.InfoD("Now Creating IAM Roles for the service account...")
	err = copier.Copy(&saAcc, saAccModel)
	if err != nil {
		return "", err
	}
	accList, err = GetAccountListv1()
	if err != nil {
		return "", err
	}
	accountId = GetPlatformAccountID(accList, accName)
	tenantId, err = GetTenantId(accountId)
	binding.RoleName = &roleName
	binding.ResourceIds = append(binding.ResourceIds, resourceId)
	iamName := "iam-" + strconv.Itoa(rand.Int())
	roles = append(roles, binding)
	actorID := *saAccModel.V1ServiceAccount.Meta.Uid
	iamRoles, err := CreatePlatformServiceAccountIamRoles(iamName, actorID, roles)
	log.FailOnError(err, "error while creating iam roles")
	log.Infof("created iam role with name %s", *iamRoles.Meta.Name)
	tokenRes, err := GenerateServiceAccountAccessToken(tenantId)
	copier.Copy(&tokenRes, generatedToken)
	jwt := generatedToken.Token
	return jwt, nil
}
