package api

import (
	"fmt"

	status "net/http"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

// BackupLocationV2 struct
type BackupLocationV2 struct {
	ApiClientV2 *platformV2.APIClient
}

// ListBackupLocations return lis of backup locations
func (backupLocation *BackupLocationV2) ListBackupLocations() ([]platformV2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientV2.BackupLocationServiceAPI
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

// GetBackupLocation get backup location model by its ID.
func (backupLocation *BackupLocationV2) GetBackupLocation(backupLocID string) (*platformV2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientV2.BackupLocationServiceAPI
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
func (backupLocation *BackupLocationV2) CreateBackupLocation(tenantID string) (*platformV2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientV2.BackupLocationServiceAPI
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
func (backupLocation *BackupLocationV2) UpdateBackupLocation(backupLocationID string) (*platformV2.V1BackupLocation, error) {
	backupLocationClient := backupLocation.ApiClientV2.BackupLocationServiceAPI
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
	backupLocationClient := backupLocation.ApiClientV2.BackupLocationServiceAPI
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
