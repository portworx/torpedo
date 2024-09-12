package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/platform/v1/serviceaccount"
	status "net/http"
)

// ListAllServiceAccounts List all Service Accounts
func (sa *PLATFORM_API_V1) ListAllServiceAccounts(listReq *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		List: []V1ServiceAccount{},
	}
	var firstPageRequest serviceaccountv1.ApiServiceAccountServiceListServiceAccountRequest
	err = copier.Copy(&firstPageRequest, listReq)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceListServiceAccountExecute(firstPageRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceListServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(saModel.ServiceAccounts, &saResponse.List)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

// GetServiceAccount return service account model.
func (sa *PLATFORM_API_V1) GetServiceAccount(saID *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		Get: V1ServiceAccount{},
	}
	var getRequest serviceaccountv1.ApiServiceAccountServiceGetServiceAccountRequest
	err = copier.Copy(&getRequest, saID)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceGetServiceAccountExecute(getRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceGetServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(saModel, &saResponse.Get)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

// CreateServiceAccount return new service account model.
func (sa *PLATFORM_API_V1) CreateServiceAccount(createSaReq *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	saResponse := PDSServiceAccountResponse{
		Create: V1ServiceAccount{},
	}
	_, saClient, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	saCreateRequest := saClient.ServiceAccountServiceCreateServiceAccount(context.Background(), createSaReq.Create.TenantId)
	saCreateRequest = saCreateRequest.V1ServiceAccount(serviceaccountv1.V1ServiceAccount{
		Meta: &serviceaccountv1.V1Meta{
			Name: createSaReq.Create.V1ServiceAccount.Meta.Name,
		},
	})
	saModel, res, err := saCreateRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ServiceAccountServiceCreateServiceAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("API response - [%+v]", saModel)
	err = utilities.CopyStruct(saModel, &saResponse.Create)
	log.Infof("API response copied - [%+v]", saResponse)
	return &saResponse, err
}

// DeleteServiceAccount delete service account and return status.
func (sa *PLATFORM_API_V1) DeleteServiceAccount(saId *PDSServiceAccountRequest) error {
	_, client, err := sa.getSAClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var delIamRequest serviceaccountv1.ApiServiceAccountServiceDeleteServiceAccountRequest
	err = copier.Copy(&delIamRequest, saId)
	if err != nil {
		return err
	}
	saModel, res, err := client.ServiceAccountServiceDeleteServiceAccountExecute(delIamRequest)
	if res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `ApiServiceAccountServiceDeleteServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return nil
}

// RegenerateServiceAccountSecret serviceAccountSecret
func (sa *PLATFORM_API_V1) RegenerateServiceAccountSecret(saId *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		RegenerateToken: GetServiceAccountTokenResponse{},
	}
	var regenIamRequest serviceaccountv1.ApiServiceAccountServiceRegenerateServiceAccountSecretRequest
	err = copier.Copy(&regenIamRequest, saId)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceRegenerateServiceAccountSecretExecute(regenIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceRegenerateServiceAccountSecretRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

// UpdateServiceAccount update existing serviceAccount
func (sa *PLATFORM_API_V1) UpdateServiceAccount(saId *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		Update: V1ServiceAccount{},
	}
	var updateIamRequest serviceaccountv1.ApiServiceAccountServiceUpdateServiceAccountRequest
	err = copier.Copy(&updateIamRequest, saId)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceUpdateServiceAccountExecute(updateIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceUpdateServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

func (sa *PLATFORM_API_V1) GenerateServiceAccountAccessToken(tokenReq *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	ctx, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		CreateToken: GetServiceAccountTokenResponse{},
	}
	var tokenIamRequest serviceaccountv1.ApiServiceAccountServiceGetAccessTokenRequest
	log.Infof("Tenant Id - [%s]", tokenReq.CreateToken.TenantId)
	tokenIamRequest = tokenIamRequest.ApiService.ServiceAccountServiceGetAccessToken(ctx, tokenReq.CreateToken.TenantId)
	tokenIamRequest = tokenIamRequest.ServiceAccountServiceGetAccessTokenBody(
		serviceaccountv1.ServiceAccountServiceGetAccessTokenBody{
			ClientId:     tokenReq.CreateToken.ServiceAccountServiceGetAccessTokenBody.ClientId,
			ClientSecret: tokenReq.CreateToken.ServiceAccountServiceGetAccessTokenBody.ClientSecret,
		},
	)
	tokenModel, res, err := client.ServiceAccountServiceGetAccessTokenExecute(tokenIamRequest)
	log.Infof("Response - [%+v]", res)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceGetAccessTokenRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", tokenModel)
	saResponse.CreateToken.Token = tokenModel.GetToken()
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}
