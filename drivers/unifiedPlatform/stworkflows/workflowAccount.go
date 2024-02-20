package stworkflows

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
)

func WorkflowCreateAndListAccounts() (map[string][]apiStructs.WorkFlowResponse, error) {
	response := utils.GetWorkflowResponseMap()

	stepName := "CreateAccount"
	startStep(stepName)
	acc, err := platformLibs.CreatePlatformAccountV1(
		fmt.Sprintf("%s-%s", envPlatformAccountName, utilities.RandomString(3)),
		fmt.Sprintf("%s-%s", envAccountDisplayName, utilities.RandomString(3)),
		fmt.Sprintf("%s-%s", envUserMailId, utilities.RandomString(3)),
	)
	if err != nil {
		return response, err
	}

	log.Infof("Created account with name %s", *acc.Meta.Name)
	response[stepName] = []apiStructs.WorkFlowResponse{acc}

	stepName = "ListAccount"
	startStep(stepName)
	accList, err := platformLibs.GetAccountListv1()
	if err != nil {
		return response, err
	}

	response[stepName] = accList

	return response, nil
}
