package api

import (
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
	ctx, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Create: V1IAM{},
	}
	iamCreateRequest := iamv1.ApiIAMServiceCreateIAMRequest{}
	iamCreateRequest = iamCreateRequest.ApiService.IAMServiceCreateIAM(ctx)
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

// GetIamRoleBindingByID return IAM RoleBinding model.
func (iam *PLATFORM_API_V1) GetIamRoleBindingByID(getRequest *IAMRequest) (*IAMResponse, error) {
	ctx, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Get: V1IAM{},
	}
	iamGetReq := iamClient.IAMServiceGetIAM(ctx, getRequest.Get.ActorId)
	if err != nil {
		return nil, err
	}
	iamModel, res, err := iamClient.IAMServiceGetIAMExecute(iamGetReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceGetIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully fetched the IAM Roles")
	err = utilities.CopyStruct(iamModel, &iamResponse.Get)
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
	ctx, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Grant: V1GrantIAMResponse{},
	}

	GrantIAMBody := iamv1.IAMServiceGrantIAMBody{
		AccountId: grantIamReq.Grant.IAMServiceGrantIAMBody.AccountId,
		ProjectId: grantIamReq.Grant.IAMServiceGrantIAMBody.ProjectId,
		TenantId:  grantIamReq.Grant.IAMServiceGrantIAMBody.TenantId,
		AccessPolicy: &iamv1.V1AccessPolicy{
			GlobalScope: []string{},
			Account:     []string{},
			Tenant:      []iamv1.V1RoleBinding{},
			Project:     []iamv1.V1RoleBinding{},
			Namespace:   []iamv1.V1RoleBinding{},
		},
	}

	// Applying all the Tenant policies
	for _, tenantPolicy := range grantIamReq.Grant.IAMServiceGrantIAMBody.Iam.Tenant {
		GrantIAMBody.AccessPolicy.Tenant = append(GrantIAMBody.AccessPolicy.Tenant, iamv1.V1RoleBinding{
			RoleName:    &tenantPolicy.RoleName,
			ResourceIds: tenantPolicy.ResourceIds,
		})
	}

	// Applying all the Project policies
	for _, projectPolicy := range grantIamReq.Grant.IAMServiceGrantIAMBody.Iam.Project {
		GrantIAMBody.AccessPolicy.Project = append(GrantIAMBody.AccessPolicy.Project, iamv1.V1RoleBinding{
			RoleName:    &projectPolicy.RoleName,
			ResourceIds: projectPolicy.ResourceIds,
		})
	}

	// Applying all the Namepsace policies
	for _, nsPolicy := range grantIamReq.Grant.IAMServiceGrantIAMBody.Iam.Namespace {
		GrantIAMBody.AccessPolicy.Namespace = append(GrantIAMBody.AccessPolicy.Namespace, iamv1.V1RoleBinding{
			RoleName:    &nsPolicy.RoleName,
			ResourceIds: nsPolicy.ResourceIds,
		})
	}

	iamTokenRequest := iamClient.IAMServiceGrantIAM(ctx, grantIamReq.Grant.IamConfigActorId).IAMServiceGrantIAMBody(GrantIAMBody)
	iamModel, res, err := iamClient.IAMServiceGrantIAMExecute(iamTokenRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceGrantIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully granted the IAM roles")
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(iamModel, &iamResponse.Grant)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}

func (iam *PLATFORM_API_V1) RevokeAccessForIAM(revokeReq *IAMRequest) (*IAMResponse, error) {
	ctx, iamClient, err := iam.getIAMClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamResponse := IAMResponse{
		Revoke: V1GrantIAMResponse{},
	}

	RevokeIAMRequest := iamv1.IAMServiceRevokeIAMBody{
		AccountId: revokeReq.Revoke.IAMServiceRevokeIAMBody.AccountId,
		ProjectId: revokeReq.Revoke.IAMServiceRevokeIAMBody.ProjectId,
		TenantId:  revokeReq.Revoke.IAMServiceRevokeIAMBody.TenantId,
		AccessPolicy: &iamv1.V1AccessPolicy{
			GlobalScope: []string{},
			Account:     []string{},
			Tenant:      []iamv1.V1RoleBinding{},
			Project:     []iamv1.V1RoleBinding{},
			Namespace:   []iamv1.V1RoleBinding{},
		},
	}

	// Applying all the Tenant policies
	for _, tenantPolicy := range revokeReq.Revoke.IAMServiceRevokeIAMBody.Iam.Tenant {
		RevokeIAMRequest.AccessPolicy.Tenant = append(RevokeIAMRequest.AccessPolicy.Tenant, iamv1.V1RoleBinding{
			RoleName:    &tenantPolicy.RoleName,
			ResourceIds: tenantPolicy.ResourceIds,
		})
	}

	// Applying all the Project policies
	for _, projectPolicy := range revokeReq.Revoke.IAMServiceRevokeIAMBody.Iam.Project {
		RevokeIAMRequest.AccessPolicy.Project = append(RevokeIAMRequest.AccessPolicy.Project, iamv1.V1RoleBinding{
			RoleName:    &projectPolicy.RoleName,
			ResourceIds: projectPolicy.ResourceIds,
		})
	}

	// Applying all the Namepsace policies
	for _, nsPolicy := range revokeReq.Revoke.IAMServiceRevokeIAMBody.Iam.Namespace {
		RevokeIAMRequest.AccessPolicy.Namespace = append(RevokeIAMRequest.AccessPolicy.Namespace, iamv1.V1RoleBinding{
			RoleName:    &nsPolicy.RoleName,
			ResourceIds: nsPolicy.ResourceIds,
		})
	}

	iamRevokeReq := iamClient.IAMServiceRevokeIAM(ctx, revokeReq.Revoke.IamConfigActorId).IAMServiceRevokeIAMBody(RevokeIAMRequest)
	iamModel, res, err := iamClient.IAMServiceRevokeIAMExecute(iamRevokeReq)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceRevokeIAMExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully revoked access to the IAM roles")
	log.Infof("Value of iam - [%v]", iamModel)
	err = utilities.CopyStruct(iamModel, &iamResponse.Revoke)
	log.Infof("Value of iam after copy - [%v]", iamResponse)
	return &iamResponse, nil
}
