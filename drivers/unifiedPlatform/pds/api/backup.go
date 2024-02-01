package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// BackupV2 struct
type BackupV2 struct {
	ApiClientV2 *pdsv2.APIClient
}

// ListBackup return list of backup models.
func (backup *BackupV2) ListBackup() ([]pdsv2.V1Backup, error) {
	backupClient := backup.ApiClientV2.BackupServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModels, res, err := backupClient.BackupServiceListBackups(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceListBackups`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModels.Backups, err
}

// ListBackupsBelongToTarget return pds backup models specific to a backup target.

// GetBackup gets pds backup model by its ID.
func (backup *BackupV2) GetBackup(backupID string) (*pdsv2.V1Backup, error) {
	backupClient := backup.ApiClientV2.BackupServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModel, res, err := backupClient.BackupServiceGetBackup(ctx, backupID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceGetBackup`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModel, err
}

// CreateBackup create adhoc/schedule backup and return the newly create backup model.

// UpdateBackup return updated backup model.

// DeleteBackupJobs delete the backup job and return the status.

// DeleteBackup delete the backup and return the status.
func (backup *BackupV2) DeleteBackup(backupID string) (*status.Response, error) {
	backupClient := backup.ApiClientV2.BackupServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := backupClient.BackupServiceDeleteBackup(ctx, backupID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceDeleteBackup`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
