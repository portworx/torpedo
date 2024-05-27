package platform

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPlatform struct {
	Accounts       map[string]map[string]string
	AdminAccountId string
	TenantId       string
}

func (platform *WorkflowPlatform) OnboardAccounts() (map[string][]automationModels.WorkFlowResponse, error) {
	resultMap := utils.GetWorkflowResponseMap()

	log.Infof("Onboarding all accounts, Results will be saved to [%s]", automationModels.CreatePlatformAccountV1)
	for accountName, accountDetails := range platform.Accounts {
		log.Infof("Onboarding Account - [%s]", accountName)
		accCreationResponse, err := platformLibs.OnboardAccount(
			accountDetails[automationModels.UserName],
			accountDetails[automationModels.UserDisplayName],
			accountDetails[automationModels.UserEmail],
		)
		if err != nil {
			return resultMap, err
		} else {
			log.Infof("Account Onboarded - UID - [%s]", *accCreationResponse.Register.Meta.Uid)
			platform.AdminAccountId = *accCreationResponse.Register.Meta.Uid
		}
	}

	return resultMap, nil
}

func (platform *WorkflowPlatform) TenantInit() (*WorkflowPlatform, error) {
	wfTenant := WorkflowTenant{
		AccountID: platform.AdminAccountId,
	}
	tenantId, err := wfTenant.GetDefaultTenantId("px-system-tenant")
	if err != nil || tenantId == "" {
		return platform, err
	}

	platform.TenantId = tenantId
	log.InfoD("TenantID is- %v", platform.TenantId)
	return platform, nil
}

func (platform *WorkflowPlatform) WhoAmI() (*automationModels.WhoamiResponse, error) {
	return platformLibs.Whoami()
}
