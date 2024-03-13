package platformLibs

import (
	"context"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/v1/serviceaccount"
)

var (
	saInputs      *automationModels.WorkFlowRequest
	saListRequest serviceaccountv1.ApiServiceAccountServiceListServiceAccountRequest
	namespaceId   string
	ServiceIdFlag bool
	SiToken       string
	SiTokenSet    string
)

// ListServiceAccountsForTenant lists all serviceAccounts for a given tenant
func ListServiceAccountsForTenant(tenantID string) ([]automationModels.WorkFlowResponse, error) {
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
func GetServiceAccountForTenant(saId, tenantId string) (*automationModels.WorkFlowResponse, error) {
	saIdModel := automationModels.WorkFlowRequest{TenantId: tenantId,
		Id: saId}
	saList, err := v2Components.Platform.GetServiceAccount(&saIdModel)
	if err != nil {
		return nil, err
	}
	return saList, nil
}

// CreateServiceAccountForRBAC creates a new service account for a given tenant
func CreateServiceAccountForRBAC(saName, tenantId string) (*automationModels.WorkFlowResponse, error) {
	saInputs.ServiceAccountRequest.V1ServiceAccount.Meta.Name = &saName
	saInputs.ServiceAccountRequest.TenantId = tenantId
	saModel, err := v2Components.Platform.CreateServiceAccount(saInputs)
	if err != nil {
		return nil, err
	}
	return saModel, nil
}

// GenerateServiceAccountAccessToken used to generate ServiceAccount JWT token
func GenerateServiceAccountAccessToken(tenantId string) (*automationModels.WorkFlowResponse, error) {
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
	log.InfoD("RBAC flag is set to- %v", ServiceIdFlag)
	return true, nil
}
