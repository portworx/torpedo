package platformLibs

import (
	"context"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/platform/v1/serviceaccount"
	"math/rand"
	"strconv"
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

func AssignRoleBindingsToUser(saName, roleName, resourceId, tenantId string) (*automationModels.WorkFlowResponse, error) {
	var (
		userModel automationModels.PDSServiceAccountRequest
		binding   automationModels.V1RoleBinding
		roles     []automationModels.V1RoleBinding
	)
	user, err := GetServiceAccFromSaName(tenantId, saName)
	err = utilities.CopyStruct(user, userModel)
	actorID := *userModel.Get.Config.ClientId
	clientSecret := *userModel.Get.Config.ClientSecret
	binding.RoleName = &roleName
	binding.ResourceIds = append(binding.ResourceIds, resourceId)
	iamName := "iam-" + strconv.Itoa(rand.Int())
	roles = append(roles, binding)
	iamRoles, err := CreatePlatformServiceAccountIamRoles(iamName, actorID, roles)
	log.FailOnError(err, "error while creating iam roles")
	log.Infof("created iam role with name %s", *iamRoles.Meta.Name)
	tokenRes, err := GenerateServiceAccountAccessToken(tenantId, actorID, clientSecret)
	return tokenRes, nil
}
