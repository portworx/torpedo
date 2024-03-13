package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publiciamapi "github.com/pure-px/apis/public/portworx/platform/iam/apiv1"
	"google.golang.org/grpc"
)

// getIamClient updates the header with bearer token and returns the new client
func (iamGrpcV1 *PlatformGrpc) getIamClient() (context.Context, publiciamapi.IAMServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var backupLocClient publiciamapi.IAMServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	backupLocClient = publiciamapi.NewIAMServiceClient(iamGrpcV1.ApiClientV1)

	return ctx, backupLocClient, token, nil
}

// ListIamRoleBindings return service identities models for a project.
func (iamGrpcV1 *PlatformGrpc) ListIamRoleBindings(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, iamClient, _, err := iamGrpcV1.getIamClient()
	iamResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *publiciamapi.ListIAMRequest
	err = copier.Copy(&firstPageRequest, listReq)
	if err != nil {
		return nil, err
	}
	iamModel, err := iamClient.ListIAM(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListIAM` call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = copier.Copy(&iamResponse, iamModel.Iam)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return iamResponse, nil
}

// CreateIamRoleBinding returns newly create IAM RoleBinding object
func (iamGrpcV1 *PlatformGrpc) CreateIamRoleBinding(createReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, iamClient, _, err := iamGrpcV1.getIamClient()
	iamResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var iamCreateRequest *publiciamapi.CreateIAMRequest
	err = copier.Copy(&iamCreateRequest, createReq)
	if err != nil {
		return nil, err
	}
	iamModel, err := iamClient.CreateIAM(ctx, iamCreateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListIAM` call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = copier.Copy(&iamResponse, iamModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

func (iamGrpcV1 *PlatformGrpc) UpdateIamRoleBindings(updateReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, iamClient, _, err := iamGrpcV1.getIamClient()
	iamResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var iamUpdateRequest *publiciamapi.UpdateIAMRequest
	err = copier.Copy(&iamUpdateRequest, updateReq)
	if err != nil {
		return nil, err
	}
	iamModel, err := iamClient.UpdateIAM(ctx, iamUpdateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListIAM` call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = copier.Copy(&iamResponse, iamModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

// GetIamRoleBindingByID return IAM RoleBinding model.
func (iamGrpcV1 *PlatformGrpc) GetIamRoleBindingByID(actorId *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, iamClient, _, err := iamGrpcV1.getIamClient()
	iamResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var iamGetRequest *publiciamapi.GetIAMRequest
	err = copier.Copy(&iamGetRequest, actorId)
	if err != nil {
		return nil, err
	}
	iamModel, err := iamClient.GetIAM(ctx, iamGetRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListIAM` call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = copier.Copy(&iamResponse, iamModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

// DeleteIamRoleBinding delete IAM RoleBinding and return status.
func (iamGrpcV1 *PlatformGrpc) DeleteIamRoleBinding(actorId *WorkFlowRequest) error {
	ctx, iamClient, _, err := iamGrpcV1.getIamClient()
	iamResponse := WorkFlowResponse{}
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var iamDelRequest *publiciamapi.DeleteIAMRequest
	err = copier.Copy(&iamDelRequest, actorId)
	if err != nil {
		return err
	}
	iamModel, err := iamClient.DeleteIAM(ctx, iamDelRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error in calling `ListIAM` call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = copier.Copy(&iamResponse, iamModel)
	if err != nil {
		return err
	}
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return nil
}

func (iamGrpcV1 *PlatformGrpc) GrantIAMRoles(grantIamReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, iamClient, _, err := iamGrpcV1.getIamClient()
	iamResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var iamGrantRequest *publiciamapi.GrantIAMRequest
	err = copier.Copy(&iamGrantRequest, grantIamReq)
	if err != nil {
		return nil, err
	}
	iamModel, err := iamClient.GrantIAM(ctx, iamGrantRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListIAM` call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = copier.Copy(&iamResponse, iamModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

func (iamGrpcV1 *PlatformGrpc) RevokeAccessForIAM(revokeReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, iamClient, _, err := iamGrpcV1.getIamClient()
	iamResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var iamRevokeRequest *publiciamapi.RevokeIAMRequest
	err = copier.Copy(&iamRevokeRequest, revokeReq)
	if err != nil {
		return nil, err
	}
	iamModel, err := iamClient.RevokeIAM(ctx, iamRevokeRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListIAM` call: %v\n", err)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = copier.Copy(&iamResponse, iamModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}
