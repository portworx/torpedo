package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// IamRoleBindingsV2 struct
type IamRoleBindingsV2 struct {
	ApiClientV2 *platformV2.APIClient
}

type NamespaceRoles struct {
	roles       platformV2.V1RoleBinding
	resourceIds []string
	roleName    string
}

// ListIamRoleBindings return service identities models for a project.
func (iam *IamRoleBindingsV2) ListIamRoleBindings() ([]platformV2.V1IAM, error) {
	iamClient := iam.ApiClientV2.IAMServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamModels, res, err := iamClient.IAMServiceListIAM(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceListIAM`: %v\n.Full HTTP response: %v", err, res)
	}
	return iamModels.Iam, nil
}

// CreateIamRoleBinding returns newly create IAM RoleBinding object
func (iam *IamRoleBindingsV2) CreateIamRoleBinding() (*platformV2.V1IAM, error) {
	iamClient := iam.ApiClientV2.IAMServiceAPI

	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamModels, res, err := iamClient.IAMServiceCreateIAM(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceCreateIAM`: %v\n.Full HTTP response: %v", err, res)
	}
	return iamModels, nil
}

func (iam *IamRoleBindingsV2) UpdateIamRoleBindings(iamMetaUid string) (*platformV2.V1IAM, error) {
	iamClient := iam.ApiClientV2.IAMServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamModels, res, err := iamClient.IAMServiceUpdateIAM(ctx, iamMetaUid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceUpdateIAM`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully updated the IAM Roles")
	return iamModels, nil
}

// GetIamRoleBindingByID return IAM RoleBinding model.
func (iam *IamRoleBindingsV2) GetIamRoleBindingByID(actorId string) (*platformV2.V1IAM, error) {
	iamClient := iam.ApiClientV2.IAMServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	iamModel, res, err := iamClient.IAMServiceGetIAM(ctx, actorId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceGetIAM`: %v\n.Full HTTP response: %v", err, res)
	}
	return iamModel, nil
}

// DeleteIamRoleBinding delete IAM RoleBinding and return status.
func (iam *IamRoleBindingsV2) DeleteIamRoleBinding(actorId string) (*status.Response, error) {
	iamClient := iam.ApiClientV2.IAMServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := iamClient.IAMServiceDeleteIAM(ctx, actorId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceDeleteIAM`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
