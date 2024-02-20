package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetIamClient updates the header with bearer token and returns the new client
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
	err = copier.Copy(&firstPageRequest, listReq)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceListServiceAccountExecute(firstPageRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceListServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel.ServiceAccounts)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
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
	err = copier.Copy(&getRequest, saID)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceGetServiceAccountExecute(getRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceGetServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
	return &saResponse, nil
}

// CreateServiceAccount return new service account model.
func (sa *PLATFORM_API_V1) CreateServiceAccount(createReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var createIamRequest platformv1.ApiServiceAccountServiceCreateServiceAccountRequest
	err = copier.Copy(&createIamRequest, createReq)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceCreateServiceAccountExecute(createIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceCreateServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
	return &saResponse, nil
}

// DeleteServiceAccount delete service account and return status.
func (sa *PLATFORM_API_V1) DeleteServiceAccount(saId *WorkFlowRequest) error {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var delIamRequest platformv1.ApiServiceAccountServiceDeleteServiceAccountRequest
	err = copier.Copy(&delIamRequest, saId)
	if err != nil {
		return err
	}
	saModel, res, err := client.ServiceAccountServiceDeleteServiceAccountExecute(delIamRequest)
	if res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `ApiServiceAccountServiceDeleteServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
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
	err = copier.Copy(&regenIamRequest, saId)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceRegenerateServiceAccountSecretExecute(regenIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceRegenerateServiceAccountSecretRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
	return &saResponse, nil
}

// UpdateServiceAccount update existing serviceAccount
func (sa *PLATFORM_API_V1) UpdateServiceAccount(saId *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var updateIamRequest platformv1.ApiServiceAccountServiceUpdateServiceAccountRequest
	err = copier.Copy(&updateIamRequest, saId)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceUpdateServiceAccountExecute(updateIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceUpdateServiceAccountRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
	return &saResponse, nil
}

func (sa *PLATFORM_API_V1) GenerateServiceAccountAccessToken(tokenReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, client, err := sa.GetSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var tokenIamRequest platformv1.ApiServiceAccountServiceGetAccessTokenRequest
	err = copier.Copy(&tokenIamRequest, tokenReq)
	if err != nil {
		return nil, err
	}
	saModel, res, err := client.ServiceAccountServiceGetAccessTokenExecute(tokenIamRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiServiceAccountServiceGetAccessTokenRequest`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
	return &saResponse, nil
}
