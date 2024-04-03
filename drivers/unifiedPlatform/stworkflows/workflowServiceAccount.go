package stworkflows

import (
	"fmt"
	_ "github.com/gobwas/glob/syntax/ast"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"strconv"
)

type WorkflowServiceAccount struct {
	UserRoles       map[string]SeviceAccount
	WorkflowProject WorkflowProject
}

type SeviceAccount struct {
	Token    string
	RoleName []string
}

const (
	ProjectAdmin = "project-admin"
	TenantAdmin  = "tenant-admin"
)

func (svcUser *WorkflowServiceAccount) CreateServiceAccount(saName string, roleName []string) (*WorkflowServiceAccount, error) {
	userDetails, err := platformLibs.CreateServiceAccountForRBAC(saName, svcUser.WorkflowProject.Platform.TenantId)
	if err != nil {
		return nil, err
	}
	log.Infof("Created service account with Name - [%s], ID - [%s]", *userDetails.Create.Meta.Name, *userDetails.Create.Meta.Uid)
	log.Infof("Assigning role bindings to the user")
	token, err := svcUser.CreateRoleBindingForUser(userDetails.Create, roleName)
	if err != nil {
		return svcUser, err
	}
	svcUser.UserRoles[saName] = SeviceAccount{
		Token:    token,
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
	log.Infof("\n\n----Switched to %s----\n\n", saName)
}

func (svcUser *WorkflowServiceAccount) SwitchToAdmin() error {
	log.Infof("\n\n----Switched to Admin User----\n\n")
	utils.RunWithRBAC = utils.RunWithRbac{
		RbacFlag: false,
	}
	return nil
}

func (svcUser *WorkflowServiceAccount) CreateRoleBindingForUser(userDetails automationModels.V1ServiceAccount, roleName []string) (string, error) {
	actorID := *userDetails.Meta.Uid
	log.Infof("Client ID - [%s], Client Secret [%s]", *userDetails.Config.ClientId, *userDetails.Config.ClientSecret)
	allRoleBindings := make(map[string][]automationModels.V1RoleBinding)
	for _, role := range roleName {
		if role == TenantAdmin {
			allRoleBindings[TenantAdmin] = append(allRoleBindings[TenantAdmin], automationModels.V1RoleBinding{
				RoleName:    TenantAdmin,
				ResourceIds: []string{svcUser.WorkflowProject.Platform.TenantId},
			})
		}

		if role == ProjectAdmin {
			allRoleBindings[ProjectAdmin] = append(allRoleBindings[ProjectAdmin], automationModels.V1RoleBinding{
				RoleName:    ProjectAdmin,
				ResourceIds: []string{svcUser.WorkflowProject.ProjectId},
			})
		}
	}
	iamName := "iam-role-for-" + *userDetails.Meta.Name + strconv.Itoa(rand.Int())
	iamRoles, err := platformLibs.CreatePlatformServiceAccountIamRoles(iamName, actorID, allRoleBindings)
	if err != nil {
		return "", err
	}
	log.Infof("created iam role with name %s", *iamRoles.Create.Meta.Name)
	tokenRes, err := platformLibs.GenerateServiceAccountAccessToken(svcUser.WorkflowProject.Platform.TenantId, *userDetails.Config.ClientId, *userDetails.Config.ClientSecret)
	if err != nil {
		return "", err
	}
	log.Infof("Token for the user - [%s]", tokenRes.CreateToken.Token)

	return tokenRes.CreateToken.Token, nil

}

func (svcUser *WorkflowServiceAccount) ValidateAccountSwitch(saName string) {

}
