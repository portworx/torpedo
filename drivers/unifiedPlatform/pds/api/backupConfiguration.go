package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// BackupConfigV2 struct
type BackupConfigV2 struct {
	ApiClientV2 *pdsv2.APIClient
}

// ListBackupConfigurations return pds backup config models.
func (backup *BackupConfigV2) ListBackupConfigurations() ([]pdsv2.V1BackupConfig, error) {
	backupClient := backup.ApiClientV2.BackupConfigServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModels, res, err := backupClient.BackupConfigServiceListBackupConfigs(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceListBackupConfigs`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModels.BackupConfigs, err
}

// GetBackupConfigurations return pds backup config model.
func (backup *BackupConfigV2) GetBackupConfigurations(backupConfigID string) (*pdsv2.V1BackupConfig, error) {
	backupClient := backup.ApiClientV2.BackupConfigServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModel, res, err := backupClient.BackupConfigServiceGetBackupConfig(ctx, backupConfigID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceGetBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModel, err
}

// CreateBackupConfigurations create backup config models.
func (backup *BackupConfigV2) CreateBackupConfigurations(projectId string) (*pdsv2.V1BackupConfig, error) {
	backupClient := backup.ApiClientV2.BackupConfigServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModel, res, err := backupClient.BackupConfigServiceCreateBackupConfig(ctx, projectId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceCreateBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModel, err
}

// UpdateBackupConfigurations return updated backup config model.
func (backup *BackupConfigV2) UpdateBackupConfigurations(projectId string) (*pdsv2.V1BackupConfig, error) {
	backupClient := backup.ApiClientV2.BackupConfigServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModel, res, err := backupClient.BackupConfigServiceUpdateBackupConfig(ctx, projectId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceUpdateBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModel, err
}

// DeleteBackupConfigurations delete the backup and return the status.
func (backup *BackupConfigV2) DeleteBackupConfigurations(backupConfigID string) (*status.Response, error) {
	backupClient := backup.ApiClientV2.BackupConfigServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := backupClient.BackupConfigServiceDeleteBackupConfig(ctx, backupConfigID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceDeleteBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
