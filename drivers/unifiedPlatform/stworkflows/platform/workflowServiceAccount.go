package platform

import (
	_ "github.com/gobwas/glob/syntax/ast"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"strconv"
)

type WorkflowServiceAccount struct {
	UserRoles        map[string]SeviceAccount
	WorkflowProjects []*WorkflowProject // Make sure all projects belong to same tenant
}

type SeviceAccount struct {
	Token       string
	RoleName    []string
	UserDetails automationModels.V1ServiceAccount
}

const (
	ProjectAdmin  = "project-admin"
	TenantAdmin   = "tenant-admin"
	ProjectWriter = "project-writer"
)

func (svcUser *WorkflowServiceAccount) CreateServiceAccount(saName string, roleName []string) (*WorkflowServiceAccount, error) {
	userDetails, err := platformLibs.CreateServiceAccountForRBAC(saName, svcUser.WorkflowProjects[0].Platform.TenantId)
	if err != nil {
		return nil, err
	}
	log.Infof("Assigning role bindings to the user")

	token, err := svcUser.CreateRoleBindingForUser(userDetails.Create, roleName)

	if err != nil {
		return svcUser, err
	}
	log.Infof("Token for [%s] is [%s]", saName, token)
	svcUser.UserRoles[saName] = SeviceAccount{
		Token:       token,
		RoleName:    roleName,
		UserDetails: userDetails.Create,
	}

	return svcUser, nil
}

func (svcUser *WorkflowServiceAccount) SwitchToServiceAccount(saName string) {
	if _, ok := svcUser.UserRoles[saName]; ok {
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

	utils.RunWithRBAC = utils.RunWithRbac{
		RbacFlag: false,
	}
	log.Infof("\n\n----Switched to Admin User----\n\n")
	return nil
}

func (svcUser *WorkflowServiceAccount) CreateRoleBindingForUser(userDetails automationModels.V1ServiceAccount, roleName []string) (string, error) {
	var allProjectIds []string

	actorID := *userDetails.Meta.Uid
	log.Infof("Client ID - [%s], Client Secret [%s]", *userDetails.Config.ClientId, *userDetails.Config.ClientSecret)

	for _, projectDetails := range svcUser.WorkflowProjects {
		allProjectIds = append(allProjectIds, projectDetails.ProjectId)
	}

	if len(roleName) > 0 {
		allRoleBindings := make(map[string][]automationModels.V1RoleBinding)
		for _, role := range roleName {
			if role == TenantAdmin {
				allRoleBindings[TenantAdmin] = append(allRoleBindings[TenantAdmin], automationModels.V1RoleBinding{
					RoleName:    TenantAdmin,
					ResourceIds: allProjectIds,
				})
			}

			if role == ProjectAdmin {
				allRoleBindings[ProjectAdmin] = append(allRoleBindings[ProjectAdmin], automationModels.V1RoleBinding{
					RoleName:    ProjectAdmin,
					ResourceIds: allProjectIds,
				})
			}

			if role == ProjectWriter {
				allRoleBindings[ProjectWriter] = append(allRoleBindings[ProjectWriter], automationModels.V1RoleBinding{
					RoleName:    ProjectWriter,
					ResourceIds: allProjectIds,
				})
			}
		}
		iamName := "iam-role-for-" + *userDetails.Meta.Name + strconv.Itoa(rand.Int())
		iamRoles, err := platformLibs.CreatePlatformServiceAccountIamRoles(iamName, actorID, allRoleBindings)
		if err != nil {
			return "", err
		}
		log.Infof("created iam role with name %s", *iamRoles.Create.Meta.Name)
	}
	tokenRes, err := platformLibs.GenerateServiceAccountAccessToken(svcUser.WorkflowProjects[0].Platform.TenantId, *userDetails.Config.ClientId, *userDetails.Config.ClientSecret)
	if err != nil {
		return "", err
	}
	log.Infof("Token for the user - [%s]", tokenRes.CreateToken.Token)

	return tokenRes.CreateToken.Token, nil

}
