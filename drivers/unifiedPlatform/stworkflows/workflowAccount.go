package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs/dataservice"
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

func DeployDataservice(ds dslibs.PDSDataService) (*apiStructs.WorkFlowResponse, error) {
	deployment, err := dslibs.DeployDataService(ds)
	if err != nil {
		return nil, err
	}
	return deployment, nil

}
