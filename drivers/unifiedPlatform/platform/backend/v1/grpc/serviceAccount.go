package grpc

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	publicsaapi "github.com/pure-px/apis/public/portworx/platform/serviceaccount/apiv1"
	"google.golang.org/grpc"
)

// getIamClient updates the header with bearer token and returns the new client
func (saGrpcV1 *PlatformGrpc) getSAClient() (context.Context, publicsaapi.ServiceAccountServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var backupLocClient publicsaapi.ServiceAccountServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	backupLocClient = publicsaapi.NewServiceAccountServiceClient(saGrpcV1.ApiClientV1)

	return ctx, backupLocClient, token, nil
}

// ListAllServiceAccounts List all Service Accounts
func (saGrpcV1 *PlatformGrpc) ListAllServiceAccounts(listReq *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		List: []V1ServiceAccount{},
	}
	var firstPageRequest *publicsaapi.ListServiceAccountRequest
	err = utilities.CopyStruct(&firstPageRequest, listReq)
	saModel, err := client.ListServiceAccount(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListServiceAccountRequest` api call: %v\n", err)
	}
	log.Infof("Value of  SA - [%v]", saModel)
	err = utilities.CopyStruct(saModel.ServiceAccounts, &saResponse.List)
	log.Infof("Value of  SA after copy - [%v]", saResponse)
	return &saResponse, nil
}

// GetServiceAccount return service account model.
func (saGrpcV1 *PlatformGrpc) GetServiceAccount(saID *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		Get: V1ServiceAccount{},
	}
	var getRequest *publicsaapi.GetServiceAccountRequest
	err = utilities.CopyStruct(&getRequest, saID)
	saModel, err := client.GetServiceAccount(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetServiceAccountRequest` api call: %v\n", err)
	}
	log.Infof("Value of  SA - [%v]", saModel)
	err = utilities.CopyStruct(saModel, &saResponse.Get)
	log.Infof("Value of  SA after copy - [%v]", saResponse)
	return &saResponse, nil
}

// CreateServiceAccount return new service account model.
func (saGrpcV1 *PlatformGrpc) CreateServiceAccount(createReq *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		Create: V1ServiceAccount{},
	}
	var createIamRequest *publicsaapi.CreateServiceAccountRequest
	err = utilities.CopyStruct(&createIamRequest, createReq)
	saModel, err := client.CreateServiceAccount(ctx, createIamRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `CreateServiceAccount` api call: %v\n", err)
	}
	log.Infof("Value of  SA - [%v]", saModel)
	err = utilities.CopyStruct(&saModel, saResponse)
	log.Infof("Value of  SA after copy - [%v]", saResponse)
	return &saResponse, nil
}

// DeleteServiceAccount delete service account and return status.
func (saGrpcV1 *PlatformGrpc) DeleteServiceAccount(saId *PDSServiceAccountRequest) error {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{}
	var delSaRequest *publicsaapi.DeleteServiceAccountRequest
	err = utilities.CopyStruct(&delSaRequest, saId)
	saModel, err := client.DeleteServiceAccount(ctx, delSaRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error in calling `DeleteServiceAccount` api call: %v\n", err)
	}
	log.Infof("Value of  SA - [%v]", saModel)
	err = utilities.CopyStruct(&saResponse, saModel)
	log.Infof("Value of  SA after copy - [%v]", saResponse)
	return nil
}

// RegenerateServiceAccountSecret serviceAccountSecret
func (saGrpcV1 *PlatformGrpc) RegenerateServiceAccountSecret(saId *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		RegenerateToken: GetServiceAccountTokenResponse{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var regenSaRequest *publicsaapi.RegenerateServiceAccountSecretRequest
	err = utilities.CopyStruct(&saResponse, saId)
	saModel, err := client.RegenerateServiceAccountSecret(ctx, regenSaRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `RegenerateServiceAccountSecret` api call: %v\n", err)
	}
	log.Infof("Value of  SA - [%v]", saModel)
	err = utilities.CopyStruct(saModel, &saResponse.RegenerateToken)
	log.Infof("Value of  SA after copy - [%v]", saResponse)
	return &saResponse, nil
}

// UpdateServiceAccount update existing serviceAccount
func (saGrpcV1 *PlatformGrpc) UpdateServiceAccount(saId *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		Update: V1ServiceAccount{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateSaRequest *publicsaapi.UpdateServiceAccountRequest
	err = utilities.CopyStruct(&saResponse, saId)
	saModel, err := client.UpdateServiceAccount(ctx, updateSaRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `UpdateServiceAccountRequest` api call: %v\n", err)
	}
	log.Infof("Value of SA - [%v]", saModel)
	err = utilities.CopyStruct(saModel, &saResponse.Update)
	log.Infof("Value of  SA after copy - [%v]", saResponse)
	return &saResponse, nil
}

func (saGrpcV1 *PlatformGrpc) GenerateServiceAccountAccessToken(tokenReq *PDSServiceAccountRequest) (*PDSServiceAccountResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := PDSServiceAccountResponse{
		CreateToken: GetServiceAccountTokenResponse{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var tokenIamRequest *publicsaapi.GetAccessTokenRequest
	err = utilities.CopyStruct(&tokenIamRequest, tokenReq)
	saModel, err := client.GetAccessToken(ctx, tokenIamRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetAccessTokenRequest` api call: %v\n", err)
	}
	err = utilities.CopyStruct(saModel, &saResponse.CreateToken)
	log.Infof("Value of  SA after copy - [%v]", saResponse)
	return &saResponse, nil
}