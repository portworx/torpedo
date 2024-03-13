package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
)

type WorkflowTenant struct {
	AccountID string
}

func (tenant *WorkflowTenant) ListTenants() (map[string][]automationModels.WorkFlowResponse, error) {
	resultMap := utils.GetWorkflowResponseMap()

	tenantsList, err := platformLibs.GetTenantListV1()
	if err != nil {
		return resultMap, err
	}
	addResultToResponse(tenantsList, automationModels.GetTenantListV1, resultMap)
	return resultMap, nil
}
