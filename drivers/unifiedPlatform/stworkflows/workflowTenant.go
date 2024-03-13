package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
)

type WorkflowTenant struct {
	AccountID string
}

func (tenant *WorkflowTenant) ListTenants() (map[string][]apiStructs.WorkFlowResponse, error) {
	resultMap := utils.GetWorkflowResponseMap()

	tenantsList, err := platformLibs.GetTenantListV1()
	if err != nil {
		return resultMap, err
	}
	addResultToResponse(tenantsList, apiStructs.GetTenantListV1, resultMap)
	return resultMap, nil
}
