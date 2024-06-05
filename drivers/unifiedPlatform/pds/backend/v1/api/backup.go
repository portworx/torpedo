package api

import (
	"fmt"
	status "net/http"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
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
func (backup *PDS_API_V1) ListBackup(listBackupRequest *automationModels.PDSBackupRequest) (*automationModels.PDSBackupResponse, error) {
	bkpResponse := automationModels.PDSBackupResponse{
		List: automationModels.PDSBackupListResponse{},
	}

	ctx, bkpClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	deploymentId := listBackupRequest.List.DeploymentId

	listBkpRequest := bkpClient.BackupServiceListBackups(ctx).DataServiceDeploymentId(deploymentId)

	if listBackupRequest.List.SortSortBy != "" {
		listBkpRequest = listBkpRequest.SortSortBy(listBackupRequest.List.SortSortBy)
	}
	if listBackupRequest.List.SortSortOrder != "" {
		listBkpRequest = listBkpRequest.SortSortOrder(listBackupRequest.List.SortSortOrder)
	}
	if listBackupRequest.List.PaginationPageNumber != "" {
		listBkpRequest = listBkpRequest.PaginationPageNumber(listBackupRequest.List.PaginationPageNumber)
	}
	if listBackupRequest.List.PaginationPageSize != "" {
		listBkpRequest = listBkpRequest.PaginationPageSize(listBackupRequest.List.PaginationPageSize)
	}

	bkpModel, res, err := bkpClient.BackupServiceListBackupsExecute(listBkpRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceListBackupsExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(bkpModel, &bkpResponse.List)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying the backup response: %v\n", err)
	}

	return &bkpResponse, nil
}

// GetBackup will fetch backup for a given deployment
func (backup *PDS_API_V1) GetBackup(getBackupRequest *automationModels.PDSBackupRequest) (*automationModels.PDSBackupResponse, error) {

	getBackupResponse := automationModels.PDSBackupResponse{
		Get: automationModels.V1Backup{},
	}

	ctx, backupClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	request := backupClient.BackupServiceGetBackup(ctx, getBackupRequest.Get.Id)
	backupModel, res, err := request.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceDeleteBackup`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(backupModel, &getBackupResponse.Get)
	if err != nil {
		return nil, err
	}

	return &getBackupResponse, err
}
