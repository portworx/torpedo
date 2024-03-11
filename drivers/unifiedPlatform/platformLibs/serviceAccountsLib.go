package platformLibs

import (
	"context"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
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
func GenerateServiceAccountAccessToken(tenantId, clientID, clientSecret string) (*apiStructs.WorkFlowResponse, error) {
	saInputs.ServiceAccountTokenRequest.TenantId = tenantId
	saInputs.ServiceAccountTokenRequest.ServiceAccountServiceGetAccessTokenBody.ClientId = &clientID
	saInputs.ServiceAccountTokenRequest.ServiceAccountServiceGetAccessTokenBody.ClientSecret = &clientSecret
	tokenModel, err := v2Components.Platform.GenerateServiceAccountAccessToken(saInputs)
	if err != nil {
		return nil, err
	}
	return tokenModel, nil
}

func GetServiceAccFromSaName(tenantId, saName string) (*apiStructs.WorkFlowResponse, error) {
	var saModel *apiStructs.WorkFlowResponse
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
