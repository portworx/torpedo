package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
)

func WorkflowCreateAndListAccounts() (map[string][]apiStructs.WorkFlowResponse, error) {
	resultMap := utils.GetWorkflowResponseMap()

	//acc, err := platformLibs.CreatePlatformAccountV1(
	//	fmt.Sprintf("%s-%s", envPlatformAccountName, utilities.RandomString(3)),
	//	fmt.Sprintf("%s-%s", envAccountDisplayName, utilities.RandomString(3)),
	//	fmt.Sprintf("%s-%s", envUserMailId, utilities.RandomString(3)),
	//)
	//if err != nil {
	//	return resultMap, err
	//}
	//
	//log.Infof("Created account with name %s", *acc.Meta.Name)
	//addResultToResponse([]apiStructs.WorkFlowResponse{acc}, CreatePlatformAccountV1, resultMap)

	accList, err := platformLibs.GetAccountListv1()
	if err != nil {
		return resultMap, err
	}

	addResultToResponse(accList, GetAccountListv1, resultMap)

	return resultMap, nil
}
