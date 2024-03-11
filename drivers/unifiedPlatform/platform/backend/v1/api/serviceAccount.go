package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

var (
	SiToken       string
	SiFlag        bool
	SaRequestBody platformv1.V1ServiceAccount
)

// GetSAClient updates the header with bearer token and returns the  client
func (sa *PLATFORM_API_V1) GetSAClient() (context.Context, *platformv1.ServiceAccountServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	sa.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	sa.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = sa.AccountID
	client := sa.ApiClientV1.ServiceAccountServiceAPI

	return ctx, client, nil
}

// ListAllServiceAccounts List all Service Accounts
func (sa *PLATFORM_API_V1) ListAllServiceAccounts(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := []WorkFlowResponse{}
	var firstPageRequest platformv1.ApiServiceAccountServiceListServiceAccountRequest
	err = utilities.CopyStruct(&firstPageRequest, listReq)
	saModel, res, err := client.ServiceAccountServiceListServiceAccountExecute(firstPageRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceListServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel.ServiceAccounts)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return saResponse, nil
}

// GetServiceAccount return service account model.
func (sa *PLATFORM_API_V1) GetServiceAccount(saID *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var getRequest platformv1.ApiServiceAccountServiceGetServiceAccountRequest
	err = utilities.CopyStruct(&getRequest, saID)
	saModel, res, err := client.ServiceAccountServiceGetServiceAccountExecute(getRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceGetServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

// CreateServiceAccount return new service account model.
func (sa *PLATFORM_API_V1) CreateServiceAccount(createSaReq *WorkFlowRequest) (*WorkFlowResponse, error) {

	saResponse := WorkFlowResponse{}
	saCreateRequest := platformv1.ApiServiceAccountServiceCreateServiceAccountRequest{}

	_, saClient, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	err = utilities.CopyStruct(&SaRequestBody, createSaReq.ServiceAccountRequest.V1ServiceAccount)
	saCreateRequest = saClient.ServiceAccountServiceCreateServiceAccount(context.Background(), createSaReq.TenantId).V1ServiceAccount(SaRequestBody)
	saModel, res, err := saClient.ServiceAccountServiceCreateServiceAccountExecute(saCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&saResponse, saModel)
	return &saResponse, err
}

// DeleteServiceAccount delete service account and return status.
func (sa *PLATFORM_API_V1) DeleteServiceAccount(saId *WorkFlowRequest) error {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var delsaRequest platformv1.ApiServiceAccountServiceDeleteServiceAccountRequest
	err = utilities.CopyStruct(&delsaRequest, saId)
	saModel, res, err := client.ServiceAccountServiceDeleteServiceAccountExecute(delsaRequest)
	if res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `ApiServiceAccountServiceDeleteServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return nil
}

// RegenerateServiceAccountSecret serviceAccountSecret
func (sa *PLATFORM_API_V1) RegenerateServiceAccountSecret(saId *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var regenIamRequest platformv1.ApiServiceAccountServiceRegenerateServiceAccountSecretRequest
	err = utilities.CopyStruct(&regenIamRequest, saId)
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
func (sa *PLATFORM_API_V1) UpdateServiceAccount(saId *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var updateSaRequest platformv1.ApiServiceAccountServiceUpdateServiceAccountRequest
	err = utilities.CopyStruct(&updateSaRequest, saId)
	saModel, res, err := client.ServiceAccountServiceUpdateServiceAccountExecute(updateSaRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceUpdateServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel)
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

func (sa *PLATFORM_API_V1) GenerateServiceAccountAccessToken(tokenReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var tokenIamRequest platformv1.ApiServiceAccountServiceGetAccessTokenRequest
	tokenIamRequest = tokenIamRequest.ApiService.ServiceAccountServiceGetAccessToken(context.Background(), tokenReq.TenantId)
	err = utilities.CopyStruct(&tokenIamRequest, tokenReq)
	tokenModel, res, err := client.ServiceAccountServiceGetAccessTokenExecute(tokenIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceGetAccessTokenRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", tokenModel)
	err = utilities.CopyStruct(&saResponse, tokenModel)
	SiFlag = true
	SiToken = tokenModel.GetToken()
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}
