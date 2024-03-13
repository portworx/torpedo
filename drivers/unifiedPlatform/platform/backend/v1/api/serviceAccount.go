package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/v1/serviceaccount"
	status "net/http"
)

var (
	SiToken       string
	SiFlag        bool
	SaRequestBody serviceaccountv1.V1ServiceAccount
)

// ListAllServiceAccounts List all Service Accounts
func (sa *PLATFORM_API_V1) ListAllServiceAccounts(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := []WorkFlowResponse{}
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
	err = copier.Copy(&saResponse, saModel.ServiceAccounts)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return saResponse, nil
}

// GetServiceAccount return service account model.
func (sa *PLATFORM_API_V1) GetServiceAccount(saID *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
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
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

// CreateServiceAccount return new service account model.
func (sa *PLATFORM_API_V1) CreateServiceAccount(createSaReq *WorkFlowRequest) (*WorkFlowResponse, error) {

	saResponse := WorkFlowResponse{}
	saCreateRequest := serviceaccountv1.ApiServiceAccountServiceCreateServiceAccountRequest{}

	_, saClient, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	err = copier.Copy(&SaRequestBody, createSaReq.ServiceAccountRequest.V1ServiceAccount)
	if err != nil {
		return nil, fmt.Errorf("Error while copying the deployment request\n")
	}

	saCreateRequest = saClient.ServiceAccountServiceCreateServiceAccount(context.Background(), createSaReq.TenantId).V1ServiceAccount(SaRequestBody)

	dsModel, res, err := saClient.ServiceAccountServiceCreateServiceAccountExecute(saCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	copier.Copy(&saResponse, dsModel)
	return &saResponse, err
}

// DeleteServiceAccount delete service account and return status.
func (sa *PLATFORM_API_V1) DeleteServiceAccount(saId *WorkFlowRequest) error {
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
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return err
	}
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return nil
}

// RegenerateServiceAccountSecret serviceAccountSecret
func (sa *PLATFORM_API_V1) RegenerateServiceAccountSecret(saId *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
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
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

// UpdateServiceAccount update existing serviceAccount
func (sa *PLATFORM_API_V1) UpdateServiceAccount(saId *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
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
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}

func (sa *PLATFORM_API_V1) GenerateServiceAccountAccessToken(tokenReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var tokenIamRequest serviceaccountv1.ApiServiceAccountServiceGetAccessTokenRequest
	tokenIamRequest = tokenIamRequest.ApiService.ServiceAccountServiceGetAccessToken(context.Background(), tokenReq.TenantId)
	err = copier.Copy(&tokenIamRequest, tokenReq)
	if err != nil {
		return nil, err
	}
	tokenModel, res, err := client.ServiceAccountServiceGetAccessTokenExecute(tokenIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceGetAccessTokenRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of ServiceAccount - [%v]", tokenModel)
	err = copier.Copy(&saResponse, tokenModel)
	if err != nil {
		return nil, err
	}
	SiFlag = true
	SiToken = tokenModel.GetToken()
	log.Infof("Value of ServiceAccount after copy - [%v]", saResponse)
	return &saResponse, nil
}
