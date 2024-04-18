package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	iamv1 "github.com/pure-px/platform-api-go-client/platform/v1/iam"
	status "net/http"
)

var IAMRequestBody iamv1.ApiIAMServiceCreateIAMRequest

// ListIamRoleBindings return service identities models for a project.
func (iam *PLATFORM_API_V1) ListIamRoleBindings(listReq *IAMRequest) (*IAMResponse, error) {
	_, iamClient, err := iam.getIAMClient()
	iamResponse := IAMResponse{
		List: ListIAM{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest iamv1.ApiIAMServiceListIAMRequest
	err = copier.Copy(&firstPageRequest, listReq.List)
	if err != nil {
		return nil, err
	}
	iamModel, res, err := iamClient.IAMServiceListIAMExecute(firstPageRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceListcloudCredations`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(iamModel, &iamResponse.List)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

// CreateIamRoleBinding returns newly create IAM RoleBinding object
func (iam *PLATFORM_API_V1) CreateIamRoleBinding(createIamReq *IAMRequest) (*IAMResponse, error) {
	_, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Create: V1IAM{},
	}
	iamCreateRequest := iamv1.ApiIAMServiceCreateIAMRequest{}
	iamCreateRequest = iamCreateRequest.ApiService.IAMServiceCreateIAM(context.Background())
	V1IAM := iamv1.V1IAM{
		Meta: &iamv1.V1Meta{
			Name: createIamReq.Create.V1IAM.Meta.Name,
		},
		Config: &iamv1.V1Config{
			ActorId:   createIamReq.Create.V1IAM.Config.ActorId,
			ActorType: createIamReq.Create.V1IAM.Config.ActorType,
			AccessPolicy: &iamv1.V1AccessPolicy{
				GlobalScope: []string{},
				Account:     []string{},
				Tenant:      []iamv1.V1RoleBinding{},
				Project:     []iamv1.V1RoleBinding{},
				Namespace:   []iamv1.V1RoleBinding{},
			},
		},
	}

	// Applying all the Tenant policies
	for _, tenantPolicy := range createIamReq.Create.V1IAM.Config.AccessPolicy.Tenant {
		V1IAM.Config.AccessPolicy.Tenant = append(V1IAM.Config.AccessPolicy.Tenant, iamv1.V1RoleBinding{
			RoleName:    &tenantPolicy.RoleName,
			ResourceIds: tenantPolicy.ResourceIds,
		})
	}

	// Applying all the Project policies
	for _, projectPolicy := range createIamReq.Create.V1IAM.Config.AccessPolicy.Project {
		V1IAM.Config.AccessPolicy.Project = append(V1IAM.Config.AccessPolicy.Project, iamv1.V1RoleBinding{
			RoleName:    &projectPolicy.RoleName,
			ResourceIds: projectPolicy.ResourceIds,
		})
	}

	// Applying all the Namepsace policies
	for _, nsPolicy := range createIamReq.Create.V1IAM.Config.AccessPolicy.Namespace {
		V1IAM.Config.AccessPolicy.Namespace = append(V1IAM.Config.AccessPolicy.Namespace, iamv1.V1RoleBinding{
			RoleName:    &nsPolicy.RoleName,
			ResourceIds: nsPolicy.ResourceIds,
		})
	}

	iamCreateRequest = iamCreateRequest.V1IAM(V1IAM)

	iamModel, res, err := iamClient.IAMServiceCreateIAMExecute(iamCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceCreateIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(iamModel, &iamResponse.Create)
	return &iamResponse, err
}

func (iam *PLATFORM_API_V1) UpdateIamRoleBindings(updateReq *IAMRequest) (*IAMResponse, error) {
	_, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Update: V1IAM{},
	}
	var iamUpdateReq iamv1.ApiIAMServiceUpdateIAMRequest
	err = copier.Copy(&iamUpdateReq, updateReq.Update)
	if err != nil {
		return nil, err
	}
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
func (iam *PLATFORM_API_V1) GetIamRoleBindingByID(getRequest *IAMRequest) (*IAMResponse, error) {
	_, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Get: V1IAM{},
	}
	var iamGetReq iamv1.ApiIAMServiceGetIAMRequest
	err = copier.Copy(&iamGetReq, getRequest.Get)
	if err != nil {
		return nil, err
	}
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
func (iam *PLATFORM_API_V1) DeleteIamRoleBinding(deleteRequest *IAMRequest) error {
	_, iamClient, err := iam.getIAMClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var iamDelReq iamv1.ApiIAMServiceDeleteIAMRequest
	err = copier.Copy(&iamDelReq, deleteRequest.Delete)
	if err != nil {
		return err
	}
	iamModel, res, err := iamClient.IAMServiceDeleteIAMExecute(iamDelReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `IAMServiceDeleteIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully DELETED the IAM Roles")
	log.Infof("Value of iam - [%v]", iamModel)
	return nil
}

func (iam *PLATFORM_API_V1) GrantIAMRoles(grantIamReq *IAMRequest) (*IAMResponse, error) {
	_, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Grant: V1GrantIAMResponse{},
	}
	var iamTokenReq iamv1.ApiIAMServiceGrantIAMRequest
	err = copier.Copy(&iamTokenReq, grantIamReq.Grant)
	if err != nil {
		return nil, err
	}
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

func (iam *PLATFORM_API_V1) RevokeAccessForIAM(revokeReq *IAMRequest) (*IAMResponse, error) {
	_, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Revoke: V1GrantIAMResponse{},
	}
	var iamRevokeReq iamv1.ApiIAMServiceRevokeIAMRequest
	err = copier.Copy(&iamRevokeReq, revokeReq)
	if err != nil {
		return nil, err
	}
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
