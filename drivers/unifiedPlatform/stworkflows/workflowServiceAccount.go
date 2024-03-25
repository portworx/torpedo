package stworkflows

import (
	"fmt"
	_ "github.com/gobwas/glob/syntax/ast"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowServiceAccount struct {
	UserRoles        map[string]SeviceAccount
	WorkflowPlatform WorkflowPlatform
}

type SeviceAccount struct {
	Token    string
	RoleName string
}

const (
	ProjectAdmin = "project-admin"
	User         = "user"
)

func (svcUser *WorkflowServiceAccount) CreateServiceAccount(accId, saName, roleName, resourceId string) (*WorkflowServiceAccount, error) {
	_, err := platformLibs.CreateServiceAccountForRBAC(saName, svcUser.WorkflowPlatform.TenantId)
	if err != nil {
		return nil, err
	}
	log.InfoD("moving on to assign roleBindings to this user")
	newToken, err := platformLibs.AssignRoleBindingsToUser(saName, roleName, resourceId, svcUser.WorkflowPlatform.TenantId)

	rbacToken := newToken.PdsServiceAccount.GetToken.Token
	svcUser.UserRoles[saName] = SeviceAccount{
		Token:    rbacToken,
		RoleName: roleName,
	}
	return svcUser, nil
}

func (svcUser *WorkflowServiceAccount) SwitchToServiceAccount(saName string) {
	if user, ok := svcUser.UserRoles[saName]; ok {
		fmt.Println("User is found", user)
		jwtToken := svcUser.UserRoles[saName]
		// check if token expired, if expired create new and update the map
		utils.RunWithRBAC = utils.RunWithRbac{
			RbacFlag:  true,
			RbacToken: jwtToken.Token,
		}
	}
}

func (svcUser *WorkflowServiceAccount) SwitchToAdmin() error {
	utils.RunWithRBAC = utils.RunWithRbac{
		RbacFlag: false,
	}
	return nil
}
