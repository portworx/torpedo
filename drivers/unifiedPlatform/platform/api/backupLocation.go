package api

import (
	"fmt"
	status "net/http"

	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
)

// BackupLocationV2 struct
type BackupLocationV2 struct {
	ApiClientv2 *pdsv2.APIClient
}

// ListBackupLocations return backup locations models
func (backupLocation *BackupLocationV2) ListBackupLocations() ([]pdsv2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientv2.BackupLocationServiceApi
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupLocationModels, res, err := backupLocationClient.BackupLocationServiceListBackupLocations(ctx).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocationServiceListBackupLocations`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupLocationModels.BackupLocations, err
}

// GetBackupLocation return backup location model.
func (backupLocation *BackupLocationV2) GetBackupLocation(backupLocID string) (*pdsv2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientv2.BackupLocationServiceApi
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceGetBackupLocation(ctx, backupLocID).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when called `BackupLocationServiceGetBackupLocation`, Full HTTP response: %v\n", res)
	}
	return backupLocationModel, err
}

// CreateBackupLocation return newly created backup location model.
func (backupLocation *BackupLocationV2) CreateBackupLocation(tenantID string) (*pdsv2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientv2.BackupLocationServiceApi
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupLocationModel, _, err := backupLocationClient.BackupLocationServiceCreateBackupLocation(ctx, tenantID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error when called `BackupLocationServiceCreateBackupLocation` to create backup target - %v", err)
	}
	return backupLocationModel, err
}

// UpdateBackupLocation return updated backup location model.
func (backupLocation *BackupLocationV2) UpdateBackupLocation(backupLocationID string) (*pdsv2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientv2.BackupLocationServiceApi
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceUpdateBackupLocation(ctx, backupLocationID).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocationServiceUpdateBackupLocation`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupLocationModel, err

}

// SyncToBackupLocation returned synced backup location model.

// DeleteBackupLocation delete backup location and return status.
func (backupLocation *BackupLocationV2) DeleteBackupLocation(backupLocationID string) (*status.Response, error) {
	backupLocationClient := backupLocation.ApiClientv2.BackupLocationServiceApi
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := backupLocationClient.BackupLocationServiceDeleteBackupLocation(ctx, backupLocationID).Execute()
	if err != nil {
		return nil, fmt.Errorf("Error when calling `BackupLocationServiceDeleteBackupLocation`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
