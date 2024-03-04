package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
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
func (saGrpcV1 *PlatformGrpc) ListAllServiceAccounts(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := []WorkFlowResponse{}
	var firstPageRequest *publicsaapi.ListServiceAccountRequest
	err = copier.Copy(&firstPageRequest, listReq)
	if err != nil {
		return nil, err
	}
	saModel, err := client.ListServiceAccount(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListServiceAccountRequest` api call: %v\n", err)
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
func (saGrpcV1 *PlatformGrpc) GetServiceAccount(saID *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var getRequest *publicsaapi.GetServiceAccountRequest
	err = copier.Copy(&getRequest, saID)
	if err != nil {
		return nil, err
	}
	saModel, err := client.GetServiceAccount(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetServiceAccountRequest` api call: %v\n", err)
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
func (saGrpcV1 *PlatformGrpc) CreateServiceAccount(createReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var createIamRequest *publicsaapi.CreateServiceAccountRequest
	err = copier.Copy(&createIamRequest, createReq)
	if err != nil {
		return nil, err
	}
	saModel, err := client.CreateServiceAccount(ctx, createIamRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `CreateServiceAccount` api call: %v\n", err)
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
func (saGrpcV1 *PlatformGrpc) DeleteServiceAccount(saId *WorkFlowRequest) error {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	var delIamRequest *publicsaapi.DeleteServiceAccountRequest
	err = copier.Copy(&delIamRequest, saId)
	if err != nil {
		return err
	}
	saModel, err := client.DeleteServiceAccount(ctx, delIamRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error in calling `DeleteServiceAccount` api call: %v\n", err)
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
func (saGrpcV1 *PlatformGrpc) RegenerateServiceAccountSecret(saId *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var regenIamRequest *publicsaapi.RegenerateServiceAccountSecretRequest
	err = copier.Copy(&regenIamRequest, saId)
	if err != nil {
		return nil, err
	}
	saModel, err := client.RegenerateServiceAccountSecret(ctx, regenIamRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `RegenerateServiceAccountSecret` api call: %v\n", err)
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
func (saGrpcV1 *PlatformGrpc) UpdateServiceAccount(saId *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateIamRequest *publicsaapi.UpdateServiceAccountRequest
	err = copier.Copy(&updateIamRequest, saId)
	if err != nil {
		return nil, err
	}
	saModel, err := client.UpdateServiceAccount(ctx, updateIamRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `UpdateServiceAccountRequest` api call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
	return &saResponse, nil
}

func (saGrpcV1 *PlatformGrpc) GenerateServiceAccountAccessToken(tokenReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, client, _, err := saGrpcV1.getSAClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	saResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var tokenIamRequest *publicsaapi.GetAccessTokenRequest
	err = copier.Copy(&tokenIamRequest, tokenReq)
	if err != nil {
		return nil, err
	}
	saModel, err := client.GetAccessToken(ctx, tokenIamRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetAccessTokenRequest` api call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", saModel)
	err = copier.Copy(&saResponse, saModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", saResponse)
	return &saResponse, nil
}
