package stworkflows

import (
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"strconv"
)

var SiToken string
var SidFlag bool

func WorkflowCreateServiceAccount(accName string) (*apiStructs.WorkFlowResponse, error) {
	accList, err := platformLibs.GetAccountListv1()
	if err != nil {
		return nil, err
	}
	accountId := platformLibs.GetPlatformAccountID(accList, accName)
	tenantId, err := platformLibs.GetTenantId(accountId)
	saName := "sa-" + strconv.Itoa(rand.Int())
	saAcc, err := platformLibs.CreateServiceAccountForRBAC(saName, tenantId)
	log.Infof("created service account with name %s", *saAcc.Meta.Name)
	return saAcc, nil
}

func WorkflowCreateIAMRolesAndGenerateJWTTokenNsLevel(accName string, saAccountModel *apiStructs.WorkFlowResponse) (string, error) {
	var (
		nsID1    []string
		nsID2    []string
		nsRoles  []apiStructs.V1RoleBinding
		binding1 apiStructs.V1RoleBinding
		binding2 apiStructs.V1RoleBinding
		saAcc    apiStructs.ServiceAccountRequest
	)
	err := copier.Copy(&saAccountModel, saAcc)
	if err != nil {
		return "Unable to get serviceAccountModel", err
	}
	targetId := ""
	accList, err := platformLibs.GetAccountListv1()
	if err != nil {
		return "", err
	}
	accountId := platformLibs.GetPlatformAccountID(accList, accName)
	tenantId, err := platformLibs.GetTenantId(accountId)
	ns1Name, ns1Id1, err := platformLibs.CreateAndFetchNamespaceId(tenantId, targetId)
	nsID1 = append(nsID1, ns1Id1)
	log.FailOnError(err, "Error while fetching namespaceID")
	log.InfoD("Namespace %v with ID1 fetched is %v ", ns1Name, nsID1)
	ns1RoleName := "namespace-admin"

	ns2Name, ns2Id2, err := platformLibs.CreateAndFetchNamespaceId(tenantId, targetId)
	nsID2 = append(nsID2, ns2Id2)
	log.FailOnError(err, "Error while fetching namespaceID")
	log.InfoD("Namespace %v with ID1 fetched is %v ", ns2Name, nsID2)
	ns2RoleName := "namespace-reader"

	binding1.ResourceIds = nsID1
	binding1.RoleName = &ns1RoleName

	binding2.ResourceIds = nsID2
	binding2.RoleName = &ns2RoleName

	nsRoles = append(nsRoles, binding1, binding2)
	iamName := "iam-" + strconv.Itoa(rand.Int())

	actorID := *saAcc.V1ServiceAccount.Meta.Uid
	iamRoles, err := platformLibs.CreatePlatformServiceAccountIamRoles(iamName, actorID, nsRoles)
	log.FailOnError(err, "error while creating iam roles")
	log.Infof("created iam role with name %s", *iamRoles.Meta.Name)
	tokenRes, err := platformLibs.GenerateServiceAccountAccessToken(tenantId)
	jwtToken := tokenRes.V1AccessToken.Token
	return jwtToken, nil
}

func SetRbacWithSAToken(flag bool, token string) {
	if flag == true {
		var rbacJwtToken apiStructs.SetRbacToken
		rbacJwtToken.SetRbac = true
		rbacJwtToken.JwtToken = token
	} else {
		SiToken = "Token for RBAC is NOT set"
	}
}
