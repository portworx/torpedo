package api

import (
	"fmt"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"

	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform"
)

// IamRoleBindingsV2 struct
type IamRoleBindingsV2 struct {
	apiClientV2 *pdsv2.APIClient
}

type NamespaceRoles struct {
	roles       pdsv2.V1RoleBinding
	resourceIds []string
	roleName    string
}

// ListIamRoleBindings return service identities models for a project.
func (iam *IamRoleBindingsV2) ListIamRoleBindings() ([]pdsv2.V1IAM, error) {
	iamClient := iam.apiClientV2.IAMServiceApi
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
func (iam *IamRoleBindingsV2) CreateIamRoleBinding() (*pdsv2.V1IAM, error) {
	iamClient := iam.apiClientV2.IAMServiceApi

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

func (iam *IamRoleBindingsV2) UpdateIamRoleBindings(iamMetaUid string) (*pdsv2.V1IAM, error) {
	iamClient := iam.apiClientV2.IAMServiceApi
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
func (iam *IamRoleBindingsV2) GetIamRoleBindingByID(actorId string) (*pdsv2.V1IAM, error) {
	iamClient := iam.apiClientV2.IAMServiceApi
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
	iamClient := iam.apiClientV2.IAMServiceApi
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
