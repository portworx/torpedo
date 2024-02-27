package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// getRestoreClient updates the header with bearer token and returns the new client
func (ds *PDSV2_API) getRestoreClient() (context.Context, *pdsv2.RestoreServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.ApiClientV2.RestoreServiceAPI

	return ctx, client, nil
}

// CreateRestore will create restore from a given backup
func (ds *PDSV2_API) CreateRestore(createRetsoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	restoreResponse := apiStructs.WorkFlowResponse{}

	_, restoreClient, err := ds.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceCreateRestoreExecute(createRetsoreRequest.Restore.Create.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceCreateRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&restoreResponse, restoreModel)
	return &restoreResponse, err

}

// ReCreateRestore will re-create restore from a given backup
func (ds *PDSV2_API) ReCreateRestore(recreateRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	restoreResponse := apiStructs.WorkFlowResponse{}

	_, restoreClient, err := ds.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceRecreateRestoreExecute(recreateRestoreRequest.Restore.ReCreate.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceRecreateRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&restoreResponse, restoreModel)
	return &restoreResponse, err

}

// GetRestore will fetch restore from existing restores
func (ds *PDSV2_API) GetRestore(getRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	restoreResponse := apiStructs.WorkFlowResponse{}

	_, restoreClient, err := ds.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceGetRestoreExecute(getRestoreRequest.Restore.Get.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceGetRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&restoreResponse, restoreModel)
	return &restoreResponse, err

}

// DeleteRestore will delete restore from existing restores
func (ds *PDSV2_API) DeleteRestore(deleteRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {

	_, restoreClient, err := ds.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	_, res, err := restoreClient.RestoreServiceDeleteRestoreExecute(deleteRestoreRequest.Restore.Delete.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceDeleteRestoreExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil, err

}

// ListRestore will list all restores
func (ds *PDSV2_API) ListRestore(listRestoreRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	restoreResponse := []apiStructs.WorkFlowResponse{}

	_, restoreClient, err := ds.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restoreModel, res, err := restoreClient.RestoreServiceListRestoresExecute(listRestoreRequest.Restore.List.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceListRestoresExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&restoreResponse, restoreModel)
	return restoreResponse, err

}
