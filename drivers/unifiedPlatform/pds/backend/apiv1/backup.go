package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// getBackupClient updates the header with bearer token and returns the new client
func (ds *PDSV2_API) getBackupClient() (context.Context, *pdsv2.BackupServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.ApiClientV2.BackupServiceAPI

	return ctx, client, nil
}

// GetBackupConfig will fetch backup for a given backup config
func (ds *PDSV2_API) GetBackup(getBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupResponse := apiStructs.WorkFlowResponse{}

	_, backupClient, err := ds.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupModel, res, err := backupClient.BackupServiceGetBackupExecute(getBackupRequest.Backup.Get.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceGetBackupExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&backupResponse, backupModel)
	return &backupResponse, err

}

// DeleteBackupConfig will delete backup for a given backup config
func (ds *PDSV2_API) DeleteBackup(deleteBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	_, backupClient, err := ds.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	_, res, err := backupClient.BackupServiceDeleteBackupExecute(deleteBackupRequest.Backup.Delete.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceDeleteBackupExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil, err
}

// ListBackupConfig will list backup for a given backup config
func (ds *PDSV2_API) ListBackup(listBackupRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	backupResponse := []apiStructs.WorkFlowResponse{}

	_, backupClient, err := ds.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupModel, res, err := backupClient.BackupServiceListBackupsExecute(listBackupRequest.Backup.List.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupServiceListBackupsExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&backupResponse, backupModel)
	return backupResponse, err

}
