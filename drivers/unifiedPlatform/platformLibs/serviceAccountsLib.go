package platformLibs

import (
	"context"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

var (
	saInputs      *apiStructs.WorkFlowRequest
	saListRequest platformv1.ApiServiceAccountServiceListServiceAccountRequest
	namespaceId   string
	ServiceIdFlag bool
	SiToken       string
	SiTokenSet    string
)

// ListServiceAccountsForTenant lists all serviceAccounts for a given tenant
func ListServiceAccountsForTenant(tenantID string) ([]apiStructs.WorkFlowResponse, error) {
	saListRequest = saListRequest.TenantId(tenantID)
	saListRequest = saListRequest.ApiService.ServiceAccountServiceListServiceAccount(context.Background())
	copier.Copy(&saInputs, saListRequest)
	saList, err := v2Components.Platform.ListAllServiceAccounts(saInputs)
	if err != nil {
		return nil, err
	}
	return saList, nil
}

// GetServiceAccountForTenant fetch ServiceAccount by its ID
func GetServiceAccountForTenant(saId, tenantId string) (*apiStructs.WorkFlowResponse, error) {
	saIdModel := apiStructs.WorkFlowRequest{TenantId: tenantId,
		Id: saId}
	saList, err := v2Components.Platform.GetServiceAccount(&saIdModel)
	if err != nil {
		return nil, err
	}
	return saList, nil
}

// CreateServiceAccountForRBAC creates a new service account for a given tenant
func CreateServiceAccountForRBAC(saName, tenantId string) (*apiStructs.WorkFlowResponse, error) {
	saInputs.ServiceAccountRequest.V1ServiceAccount.Meta.Name = &saName
	saInputs.ServiceAccountRequest.TenantId = tenantId
	saModel, err := v2Components.Platform.CreateServiceAccount(saInputs)
	if err != nil {
		return nil, err
	}
	return saModel, nil
}

// GenerateServiceAccountAccessToken used to generate ServiceAccount JWT token
func GenerateServiceAccountAccessToken(tenantId string) (*apiStructs.WorkFlowResponse, error) {
	saInputs.ServiceAccountTokenRequest.TenantId = tenantId
	tokenModel, err := v2Components.Platform.GenerateServiceAccountAccessToken(saInputs)
	if err != nil {
		return nil, err
	}
	return tokenModel, nil
}

// SetRbacWithSAToken used by testcases to toggle between access token
func SetRbacWithSAToken(value bool, token string) (bool, error) {
	if value == true {
		ServiceIdFlag = true
		SiTokenSet = token
	} else {
		ServiceIdFlag = false
	}
	log.InfoD("Successfully updated Infra params for ServiceIdentity and RBAC test")
	log.InfoD("ServiceIdentity flag is set to- %v", ServiceIdFlag)
	return true, nil
}
