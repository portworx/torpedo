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
func (backup *PDS_API_V1) ListBackup(listBackupConfigRequest *automationModels.PDSBackupRequest) ([]automationModels.PDSBackupResponse, error) {
	bkpResponse := []automationModels.PDSBackupResponse{}

	ctx, bkpClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	backupConfigId := listBackupConfigRequest.List.BackupConfigId
	namespaceId := listBackupConfigRequest.List.NamespaceId
	targetClusterId := listBackupConfigRequest.List.TargetClusterId
	deploymentId := listBackupConfigRequest.List.DeploymentId

	listBkpRequest := bkpClient.BackupServiceListBackups(ctx).BackupConfigId(backupConfigId).TargetClusterId(targetClusterId).NamespaceId(namespaceId).DeploymentId(deploymentId)

	bkpModel, res, err := bkpClient.BackupServiceListBackupsExecute(listBkpRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(bkpModel.Backups, &bkpResponse)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying the backup response: %v\n", err)
	}

	return bkpResponse, nil
}

// GetBackup will fetch backup for a given deployment
func (backup *PDS_API_V1) GetBackup(getBackupConfigRequest *automationModels.PDSBackupRequest) (*automationModels.PDSBackupResponse, error) {

	response := automationModels.PDSBackupResponse{}

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
