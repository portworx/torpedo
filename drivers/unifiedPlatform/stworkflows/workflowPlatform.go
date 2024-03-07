package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPlatform struct {
	Accounts map[string]map[string]string
}

func (platform *WorkflowPlatform) CreateAccounts() (map[string][]apiStructs.WorkFlowResponse, error) {
	resultMap := utils.GetWorkflowResponseMap()

	log.Infof("Creating all accounts, Results will be saved to [%s]", apiStructs.CreatePlatformAccountV1)
	for accountName, accountDetails := range platform.Accounts {
		log.Infof("Creating Account - [%s]", accountName)
		accCreationResponse, err := platformLibs.CreatePlatformAccountV1(
			accountDetails[apiStructs.UserName],
			accountDetails[apiStructs.UserDisplayName],
			accountDetails[apiStructs.UserEmail],
		)
		if err != nil {
			return resultMap, err
		} else {
			log.Infof("Account Created - UID - [%s]", *accCreationResponse.Meta.Uid)
			addResultToResponse([]apiStructs.WorkFlowResponse{accCreationResponse}, apiStructs.CreatePlatformAccountV1, resultMap)
		}
	}

	log.Infof("Verifying if the accounts are created successfully")
	for _, accountResponse := range resultMap[apiStructs.CreatePlatformAccountV1] {
		accountResponse, err := platformLibs.GetAccount(*accountResponse.Meta.Uid)
		if err != nil {
			log.Infof("Unable to get account details for [%s], UID - [%s]", *accountResponse.Meta.Name, *accountResponse.Meta.Uid)
			return resultMap, err
		} else {
			log.Infof("Account found - [%s], UID - [%s]", *accountResponse.Meta.Name, *accountResponse.Meta.Uid)
		}
	}

	return resultMap, nil
}
