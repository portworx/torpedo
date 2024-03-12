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

var IAMRequestBody platformv1.ApiIAMServiceCreateIAMRequest

// GetIamClient updates the header with bearer token and returns the  client
func (iam *PLATFORM_API_V1) GetIamClient() (context.Context, *platformv1.IAMServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	iam.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	iam.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = iam.AccountID
	client := iam.ApiClientV1.IAMServiceAPI

	return ctx, client, nil
}

// ListIamRoleBindings return service identities models for a project.
func (iam *PLATFORM_API_V1) ListIamRoleBindings(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	_, iamClient, err := iam.GetIamClient()
	iamResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest platformv1.ApiIAMServiceListIAMRequest
	err = utilities.CopyStruct(&firstPageRequest, listReq)
	iamModel, res, err := iamClient.IAMServiceListIAMExecute(firstPageRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceListcloudCredations`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(&iamResponse, iamModel.Iam)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return iamResponse, nil
}

// CreateIamRoleBinding returns newly create IAM RoleBinding object
func (iam *PLATFORM_API_V1) CreateIamRoleBinding(createIamReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, iamClient, err := iam.GetIamClient()
	iamResponse := WorkFlowResponse{}
	iamCreateRequest := platformv1.ApiIAMServiceCreateIAMRequest{}
	iamCreateRequest = iamCreateRequest.ApiService.IAMServiceCreateIAM(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	err = utilities.CopyStruct(&IAMRequestBody, createIamReq.Iam)
	iamModel, res, err := iamClient.IAMServiceCreateIAMExecute(iamCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&iamResponse, iamModel)
	return &iamResponse, err
}

func (iam *PLATFORM_API_V1) UpdateIamRoleBindings(updateReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, iamClient, err := iam.GetIamClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := WorkFlowResponse{}
	var iamUpdateReq platformv1.ApiIAMServiceUpdateIAMRequest
	err = utilities.CopyStruct(&iamUpdateReq, updateReq)
	iamModel, res, err := iamClient.IAMServiceUpdateIAMExecute(iamUpdateReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceUpdateIAM`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully updated the IAM Roles")
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(&iamResponse, iamModel)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

// GetIamRoleBindingByID return IAM RoleBinding model.
func (iam *PLATFORM_API_V1) GetIamRoleBindingByID(actorId *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, iamClient, err := iam.GetIamClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := WorkFlowResponse{}
	var iamGetReq platformv1.ApiIAMServiceGetIAMRequest
	err = utilities.CopyStruct(&iamGetReq, actorId)
	iamModel, res, err := iamClient.IAMServiceGetIAMExecute(iamGetReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceGetIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully fetched the IAM Roles")
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(&iamResponse, iamModel)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

// DeleteIamRoleBinding delete IAM RoleBinding and return status.
func (iam *PLATFORM_API_V1) DeleteIamRoleBinding(actorId *WorkFlowRequest) error {
	_, iamClient, err := iam.GetIamClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := WorkFlowResponse{}
	var iamDelReq platformv1.ApiIAMServiceDeleteIAMRequest
	err = utilities.CopyStruct(&iamDelReq, actorId)
	iamModel, res, err := iamClient.IAMServiceDeleteIAMExecute(iamDelReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `IAMServiceDeleteIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully DELETED the IAM Roles")
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(&iamResponse, iamModel)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return nil
}

func (iam *PLATFORM_API_V1) GrantIAMRoles(grantIamReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, iamClient, err := iam.GetIamClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := WorkFlowResponse{}
	var iamTokenReq platformv1.ApiIAMServiceGrantIAMRequest
	err = utilities.CopyStruct(&iamTokenReq, grantIamReq)
	iamModel, res, err := iamClient.IAMServiceGrantIAMExecute(iamTokenReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceGrantIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully granted the IAM roles")
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(&iamResponse, iamModel)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

func (iam *PLATFORM_API_V1) RevokeAccessForIAM(revokeReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, iamClient, err := iam.GetIamClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := WorkFlowResponse{}
	var iamRevokeReq platformv1.ApiIAMServiceRevokeIAMRequest
	err = utilities.CopyStruct(&iamRevokeReq, revokeReq)
	iamModel, res, err := iamClient.IAMServiceRevokeIAMExecute(iamRevokeReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceGrantIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully revoked access to the IAM roles")
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(&iamResponse, iamModel)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}
