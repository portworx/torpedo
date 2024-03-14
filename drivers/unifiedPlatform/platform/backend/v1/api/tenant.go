package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// ListTenants return pds tenants models.
func (tenant *PLATFORM_API_V1) ListTenants() ([]WorkFlowResponse, error) {
	tenantsResponse := []WorkFlowResponse{}
	ctx, tenantClient, err := tenant.getTenantClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	//tenantsModel, res, err := tenantClient.TenantServiceListTenantsExecute(req)
	tenantsModel, res, err := tenantClient.TenantServiceListTenants(ctx).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TenantServiceListTenants2`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of tenants - [%v]", tenantsModel)
	err = copier.Copy(&tenantsResponse, tenantsModel.Tenants)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of accounts after copy - [%v]", tenantsResponse)

	return tenantsResponse, nil
}
