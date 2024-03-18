package platformLibs

import (
	"context"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/platform/v1/serviceaccount"
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
	err = utilities.CopyStruct(&saInputs, saListRequest)
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
	saInputs.ServiceAccountRequest.Create.V1ServiceAccount.Meta.Name = &saName
	saInputs.ServiceAccountRequest.Create.TenantId = tenantId
	saModel, err := v2Components.Platform.CreateServiceAccount(saInputs)
	if err != nil {
		return nil, err
	}
	return saModel, nil
}

// GenerateServiceAccountAccessToken used to generate ServiceAccount JWT token
func GenerateServiceAccountAccessToken(tenantId, clientID, clientSecret string) (*automationModels.WorkFlowResponse, error) {
	saInputs.ServiceAccountRequest.CreateToken.TenantId = tenantId
	saInputs.ServiceAccountRequest.CreateToken.ServiceAccountServiceGetAccessTokenBody.ClientId = &clientID
	saInputs.ServiceAccountRequest.CreateToken.ServiceAccountServiceGetAccessTokenBody.ClientSecret = &clientSecret
	tokenModel, err := v2Components.Platform.GenerateServiceAccountAccessToken(saInputs)
	if err != nil {
		return nil, err
	}
	return tokenModel, nil
}

func GetServiceAccFromSaName(tenantId, saName string) (*automationModels.WorkFlowResponse, error) {
	var saModel *automationModels.WorkFlowResponse
	saList, err := ListServiceAccountsForTenant(tenantId)
	if err != nil {
		return nil, err
	}
	for _, sa := range saList {
		if *sa.Meta.Name == saName {
			saModel = &sa
		}
	}
	return saModel, nil
}
