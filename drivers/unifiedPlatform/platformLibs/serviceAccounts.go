package platformLibs

import (
	"context"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/platform/v1/serviceaccount"
)

var (
	saInputs      *automationModels.PDSServiceAccountRequest
	saListRequest serviceaccountv1.ApiServiceAccountServiceListServiceAccountRequest
)

// ListServiceAccountsForTenant lists all serviceAccounts for a given tenant
func ListServiceAccountsForTenant(tenantID string) (*automationModels.PDSServiceAccountResponse, error) {
	saListRequest = saListRequest.TenantId(tenantID)
	saListRequest = saListRequest.ApiService.ServiceAccountServiceListServiceAccount(context.Background())
	err = utilities.CopyStruct(saListRequest, &saInputs)
	saList, err := v2Components.Platform.ListAllServiceAccounts(saInputs)
	if err != nil {
		return nil, err
	}
	return saList, nil
}

// GetServiceAccountForTenant fetch ServiceAccount by its ID
func GetServiceAccountForTenant(saId, tenantId string) (*automationModels.PDSServiceAccountResponse, error) {
	saIdModel := automationModels.PDSServiceAccountRequest{
		Get: automationModels.GetServiceAccount{
			TenantId: tenantId,
			Id:       saId,
		}}
	saList, err := v2Components.Platform.GetServiceAccount(&saIdModel)
	if err != nil {
		return nil, err
	}
	return saList, nil
}

// CreateServiceAccountForRBAC creates a new service account for a given tenant
func CreateServiceAccountForRBAC(saName, tenantId string) (*automationModels.PDSServiceAccountResponse, error) {
	log.Infof("SA Name - [%s]", saName)
	log.Infof("Tenant Id - [%s]", tenantId)
	saInputs := automationModels.PDSServiceAccountRequest{
		Create: automationModels.CreateServiceAccounts{
			V1ServiceAccount: automationModels.V1ServiceAccount{
				Meta: automationModels.Meta{
					Name: &saName,
				},
			},
			TenantId: tenantId,
		},
	}
	saInputs.Create.V1ServiceAccount.Meta.Name = &saName
	saInputs.Create.TenantId = tenantId
	saModel, err := v2Components.Platform.CreateServiceAccount(&saInputs)
	if err != nil {
		return nil, err
	}
	return saModel, nil
}

// GenerateServiceAccountAccessToken used to generate ServiceAccount JWT token
func GenerateServiceAccountAccessToken(tenantId string, clientID string, clientSecret string) (*automationModels.PDSServiceAccountResponse, error) {
	saInputs = &automationModels.PDSServiceAccountRequest{
		CreateToken: automationModels.CreatePdsServiceAccountToken{
			TenantId: tenantId,
			ServiceAccountServiceGetAccessTokenBody: automationModels.ServiceAccountServiceGetAccessTokenBody{
				ClientId:     &clientID,
				ClientSecret: &clientSecret,
			},
		},
	}
	tokenModel, err := v2Components.Platform.GenerateServiceAccountAccessToken(saInputs)
	if err != nil {
		return nil, err
	}
	return tokenModel, nil
}

func GetServiceAccFromSaName(tenantId, saName string) (*automationModels.V1ServiceAccount, error) {
	var saModel *automationModels.V1ServiceAccount
	saList, err := ListServiceAccountsForTenant(tenantId)
	if err != nil {
		return nil, err
	}
	for _, sa := range saList.List {
		if *sa.Meta.Name == saName {
			saModel = &sa
		}
	}
	return saModel, nil
}
