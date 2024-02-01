package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// RestoreV2 struct
type RestoreV2 struct {
	ApiClientV2 *pdsv2.APIClient
}

// ListRestores return restore models.
func (restore *RestoreV2) ListRestores() ([]pdsv2.V1Restore, error) {
	dsClient := restore.ApiClientV2.RestoreServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := dsClient.RestoreServiceListRestores(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceListRestores`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel.Restores, err
}

// GetRestoreById get restore model by its ID.
func (restore *RestoreV2) GetRestoreById(restoreId string) (*pdsv2.V1Restore, error) {
	dsClient := restore.ApiClientV2.RestoreServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := dsClient.RestoreServiceGetRestore(ctx, restoreId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceGetRestore`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel, err
}

// CreateRestore creates new restore model.
func (restore *RestoreV2) CreateRestore(namespaceId string) (*pdsv2.V1Restore, error) {
	dsClient := restore.ApiClientV2.RestoreServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := dsClient.RestoreServiceCreateRestore(ctx, namespaceId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceCreateRestore`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel, err
}

// RecreateRestore creates new restore model.
func (restore *RestoreV2) RecreateRestore(restoreId string) (*pdsv2.V1Restore, error) {
	dsClient := restore.ApiClientV2.RestoreServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	restoreModel, res, err := dsClient.RestoreServiceRecreateRestore(ctx, restoreId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceRecreateRestore`: %v\n.Full HTTP response: %v", err, res)
	}
	return restoreModel, err
}
