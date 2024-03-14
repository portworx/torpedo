package stworkflows

import (
	"fmt"
	_ "github.com/gobwas/glob/syntax/ast"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
)

type UserWithRbac struct {
	UserRoles map[string]string
}

func (svcUser *UserWithRbac) CreateNewPdsUser(accId, saName, roleName, resourceId string) (*apiStructs.WorkFlowResponse, error) {
	saModel, err := platformLibs.CreateUser(saName, accId)
	if err != nil {
		return nil, err
	}
	log.InfoD("moving on to assign roleBindings to this user")
	newToken, err := platformLibs.AssignRoleBindingsToUser(saName, roleName, resourceId, accId)

	rbacToken := newToken.PdsServiceAccount.GetToken.Token
	svcUser.UserRoles[saName] = rbacToken
	return saModel, nil
}

func (svcUser *UserWithRbac) SwitchPdsUser(saName string) {
	// Check if a user exists
	defer func() {
		utils.RunWithRBAC = utils.RunWithRbac{
			RbacFlag: false,
		}
	}()
	if user, ok := svcUser.UserRoles[saName]; ok {
		fmt.Println("User is found", user)
		jwtToken := svcUser.UserRoles[saName]
		// check if token expired, if expired create new and update the map
		utils.RunWithRBAC = utils.RunWithRbac{
			RbacFlag:  true,
			RbacToken: jwtToken,
		}
	}
}
