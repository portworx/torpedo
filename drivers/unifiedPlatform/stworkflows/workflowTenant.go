package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
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

func (tenant *WorkflowTenant) GetDefaultTenantId() (string, error) {
	var tenantId string
	tenantList, err := tenant.ListTenants()
	if err != nil {
		return "", err
	}
	for _, thistenant := range tenantList[automationModels.GetTenantListV1] {
		log.Infof("Available tenant's %s under the account id %s", *thistenant.Meta.Name, tenant.AccountID)
		tenantId = *thistenant.Meta.Uid
		break
	}
	return tenantId, nil
}
