package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	status "net/http"
)

// getRestoreClient updates the header with bearer token and returns the new client
func (restore *PDSV2_API) getRestoreClient() (context.Context, *pdsv2.RestoreServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	restore.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	restore.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = restore.AccountID
	client := restore.ApiClientV2.RestoreServiceAPI

	return ctx, client, nil
}

// CreateRestore will create restore from a given backup
func (restore *PDSV2_API) CreateRestore(createRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	restoreResponse := apiStructs.WorkFlowResponse{}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreCreateRequest := restoreClient.RestoreServiceCreateRestore(ctx, createRestoreRequest.Restore.V1.Create.NamespaceId)
	err = utilities.CopyStruct(restoreCreateRequest, createRestoreRequest.Restore.V1.Create)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceCreateRestoreExecute(restoreCreateRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceCreateRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(&restoreResponse, restoreModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return &restoreResponse, err

}

// ReCreateRestore will re-create restore from a given backup
func (restore *PDSV2_API) ReCreateRestore(recreateRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	restoreResponse := apiStructs.WorkFlowResponse{}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreReCreateRequest := restoreClient.RestoreServiceRecreateRestore(ctx, recreateRestoreRequest.Restore.V1.ReCreate.Id)
	err = utilities.CopyStruct(restoreReCreateRequest, recreateRestoreRequest.Restore.V1.ReCreate)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceRecreateRestoreExecute(restoreReCreateRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceRecreateRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(&restoreResponse, restoreModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return &restoreResponse, err

}

// GetRestore will fetch restore from existing restores
func (restore *PDSV2_API) GetRestore(getRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	restoreResponse := apiStructs.WorkFlowResponse{}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreGetRequest := restoreClient.RestoreServiceGetRestore(ctx, getRestoreRequest.Restore.V1.Get.Id)
	err = utilities.CopyStruct(restoreGetRequest, getRestoreRequest.Restore.V1.Get)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceGetRestoreExecute(restoreGetRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceGetRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(&restoreResponse, restoreModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return &restoreResponse, err

}

// DeleteRestore will delete restore from existing restores
func (restore *PDSV2_API) DeleteRestore(deleteRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreDeleteRequest := restoreClient.RestoreServiceDeleteRestore(ctx, deleteRestoreRequest.Restore.V1.Delete.Id)
	err = utilities.CopyStruct(restoreDeleteRequest, deleteRestoreRequest.Restore.V1.Delete)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	_, res, err := restoreClient.RestoreServiceDeleteRestoreExecute(restoreDeleteRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceDeleteRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil, err

}

// ListRestore will list all restores
func (restore *PDSV2_API) ListRestore(listRestoreRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	restoreResponse := []apiStructs.WorkFlowResponse{}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreListRequest := restoreClient.RestoreServiceListRestores(ctx)
	err = utilities.CopyStruct(restoreListRequest, listRestoreRequest.Restore.V1.List)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceListRestoresExecute(restoreListRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceListRestoresExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(restoreResponse, restoreModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return restoreResponse, err

}
