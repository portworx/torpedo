package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowTenant struct {
	AccountID string
	Tenants   []Tenants
}

type Tenants struct {
	TenantName string
	TenantId   string
}

func (tenant *WorkflowTenant) ListTenants() (*WorkflowTenant, error) {
	tenantsList, err := platformLibs.GetTenantListV1()
	if err != nil {
		return nil, err
	}
	tenant.Tenants = make([]Tenants, len(tenantsList))
	for index, ten := range tenantsList {
		tenant.Tenants[index].TenantName = *ten.Meta.Name
		tenant.Tenants[index].TenantId = *ten.Meta.Uid
	}

	return tenant, nil
}

func (tenant *WorkflowTenant) GetDefaultTenantId(tenantName string) (string, error) {
	var tenantId string
	tenantList, err := tenant.ListTenants()
	if err != nil {
		return "", err
	}

	for _, tenant := range tenantList.Tenants {
		log.Infof("Name [%s]", tenant.TenantName)
		log.Infof("Id [%s]", tenant.TenantId)
		if tenant.TenantName == tenantName {
			tenantId = tenant.TenantId
		}
	}

	return tenantId, nil
}
