package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPlatform struct {
	Accounts       map[string]map[string]string
	AdminAccountId string
	TenantId       string
}

func (platform *WorkflowPlatform) OnboardAccounts() (map[string][]apiStructs.WorkFlowResponse, error) {
	resultMap := utils.GetWorkflowResponseMap()

	log.Infof("Onboarding all accounts, Results will be saved to [%s]", apiStructs.CreatePlatformAccountV1)
	for accountName, accountDetails := range platform.Accounts {
		log.Infof("Onboarding Account - [%s]", accountName)
		accCreationResponse, err := platformLibs.OnboardAccount(
			accountDetails[apiStructs.UserName],
			accountDetails[apiStructs.UserDisplayName],
			accountDetails[apiStructs.UserEmail],
		)
		if err != nil {
			return resultMap, err
		} else {
			log.Infof("Account Onboarded - UID - [%s]", *accCreationResponse.OnboardAccount.Meta.Uid)
			platform.AdminAccountId = *accCreationResponse.OnboardAccount.Meta.Uid
			addResultToResponse([]apiStructs.WorkFlowResponse{*accCreationResponse}, apiStructs.CreatePlatformAccountV1, resultMap)
		}
	}

	return resultMap, nil
}

func (platform *WorkflowPlatform) TenantInit() (*WorkflowPlatform, error) {
	wfTenant := WorkflowTenant{
		AccountID: platform.AdminAccountId,
	}
	tenantId, err := wfTenant.GetDefaultTenantId()
	if err != nil {
		return platform, err
	}
	platform.TenantId = tenantId

	return platform, nil
}
