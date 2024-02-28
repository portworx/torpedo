package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	status "net/http"
)

// getBackupClient updates the header with bearer token and returns the new client
func (backup *PDSV2_API) getBackupClient() (context.Context, *pdsv2.BackupServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	backup.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	backup.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = backup.AccountID
	client := backup.ApiClientV2.BackupServiceAPI

	return ctx, client, nil
}

// GetBackupConfig will fetch backup for a given backup config
func (backup *PDSV2_API) GetBackup(getBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupResponse := apiStructs.WorkFlowResponse{}

	ctx, backupClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupGetRequest := backupClient.BackupServiceGetBackup(ctx, getBackupRequest.Id)
	err = utilities.CopyStruct(backupGetRequest, getBackupRequest.Backup.V1.Get)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	backupModel, res, err := backupClient.BackupServiceGetBackupExecute(backupGetRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceGetBackupExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&backupResponse, backupModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return &backupResponse, err

}

// DeleteBackupConfig will delete backup for a given backup config
func (backup *PDSV2_API) DeleteBackup(deleteBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	ctx, backupClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupDeleteRequest := backupClient.BackupServiceDeleteBackup(ctx, deleteBackupRequest.Id)
	err = utilities.CopyStruct(backupDeleteRequest, deleteBackupRequest.Backup.V1.Delete)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	_, res, err := backupClient.BackupServiceDeleteBackupExecute(backupDeleteRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceDeleteBackupExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil, err
}

// ListBackupConfig will list backup for a given backup config
func (backup *PDSV2_API) ListBackup(listBackupRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	backupResponse := []apiStructs.WorkFlowResponse{}

	ctx, backupClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupListRequest := backupClient.BackupServiceListBackups(ctx)
	err = utilities.CopyStruct(backupListRequest, listBackupRequest.Backup.V1.List)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	backupModel, res, err := backupClient.BackupServiceListBackupsExecute(backupListRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceListBackupsExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(backupResponse, backupModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return backupResponse, err

}
