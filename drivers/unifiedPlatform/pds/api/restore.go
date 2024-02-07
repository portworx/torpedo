package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// RestoreV2 struct
type RestoreV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (restore *RestoreV2) GetClient() (context.Context, *pdsv2.RestoreServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	restore.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	restore.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = restore.AccountID
	client := restore.ApiClientV2.RestoreServiceAPI
	return ctx, client, nil
}

// ListRestores return restore models.
func (restore *RestoreV2) ListRestores() ([]pdsv2.V1Restore, error) {
	ctx, restoreClient, err := restore.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := restoreClient.RestoreServiceListRestores(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceListRestores`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel.Restores, err
}

// GetRestoreById get restore model by its ID.
func (restore *RestoreV2) GetRestoreById(restoreId string) (*pdsv2.V1Restore, error) {
	ctx, restoreClient, err := restore.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := restoreClient.RestoreServiceGetRestore(ctx, restoreId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceGetRestore`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel, err
}

// CreateRestore creates new restore model.
func (restore *RestoreV2) CreateRestore(namespaceId string) (*pdsv2.V1Restore, error) {
	ctx, restoreClient, err := restore.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := restoreClient.RestoreServiceCreateRestore(ctx, namespaceId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceCreateRestore`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel, err
}

// RecreateRestore creates new restore model.
func (restore *RestoreV2) RecreateRestore(restoreId string) (*pdsv2.V1Restore, error) {
	ctx, restoreClient, err := restore.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := restoreClient.RestoreServiceRecreateRestore(ctx, restoreId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceRecreateRestore`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel, err
}
