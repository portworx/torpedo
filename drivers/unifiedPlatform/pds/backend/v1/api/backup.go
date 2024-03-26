package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	status "net/http"
)

// DeleteBackup will delete backup for a given deployment
func (backup *PDS_API_V1) DeleteBackup(deleteBackupRequest *automationModels.PDSBackupRequest) error {

	ctx, backupClient, err := backup.getBackupClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	deleteRequest := backupClient.BackupServiceDeleteBackup(ctx, deleteBackupRequest.Delete.Id)
	_, res, err := deleteRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `BackupServiceDeleteBackup`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil
}

// ListBackup will list backup for a given deployment
func (backup *PDS_API_V1) ListBackup(listBackupConfigRequest *automationModels.PDSBackupRequest) (*automationModels.PDSBackupResponse, error) {
	response := automationModels.PDSBackupResponse{
		Get: automationModels.V1Backup{},
	}

	ctx, backupClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	listBackupRequest := backupClient.BackupServiceListBackups(ctx)

	listBackupRequest = listBackupRequest.NamespaceId(listBackupConfigRequest.List.NamespaceId)
	listBackupRequest = listBackupRequest.DeploymentId(listBackupConfigRequest.List.DeploymentId)
	listBackupRequest = listBackupRequest.BackupConfigId(listBackupConfigRequest.List.BackupConfigId)
	listBackupRequest = listBackupRequest.TargetClusterId(listBackupConfigRequest.List.TargetClusterId)

	backupModel, res, err := listBackupRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceDeleteBackup`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(backupModel, &response)
	if err != nil {
		return nil, err
	}

	return &response, err
}

// GetBackup will fetch backup for a given deployment
func (backup *PDS_API_V1) GetBackup(getBackupConfigRequest *automationModels.PDSBackupRequest) (*automationModels.PDSBackupResponse, error) {

	response := automationModels.PDSBackupResponse{
		List: automationModels.PDSBackupListResponse{},
	}

	ctx, backupClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	getBackupRequest := backupClient.BackupServiceGetBackup(ctx, getBackupConfigRequest.Get.Id)
	backupModel, res, err := getBackupRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceDeleteBackup`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(backupModel, &response)
	if err != nil {
		return nil, err
	}

	return &response, err
}
