package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
)

func WorkflowListTenants(accountID string) (map[string][]apiStructs.WorkFlowResponse, error) {
	resultMap := utils.GetWorkflowResponseMap()

	tenantsList, err := platformLibs.GetTenantListV1(accountID)
	if err != nil {
		return resultMap, err
	}
	addResultToResponse(tenantsList, GetTenantListV1, resultMap)
	return resultMap, nil
}
