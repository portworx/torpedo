package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// DeleteBackup will delete backup for a given deployment
func (backup *PDS_API_V1) DeleteBackup(deleteBackupRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("DeleteBackup is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// ListBackup will list backup for a given deployment
func (backup *PDS_API_V1) ListBackup(listBackupConfigRequest *automationModels.WorkFlowRequest) ([]automationModels.PDSBackupResponse, error) {
	bkpResponse := []automationModels.PDSBackupResponse{}

	ctx, bkpClient, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	backupConfigId := listBackupConfigRequest.Backup.List.BackupConfigId
	namespaceId := listBackupConfigRequest.Backup.List.NamespaceId
	targetClusterId := listBackupConfigRequest.Backup.List.TargetClusterId
	deploymentId := listBackupConfigRequest.Backup.List.DeploymentId

	listBkpRequest := bkpClient.BackupServiceListBackups(ctx).BackupConfigId(backupConfigId).TargetClusterId(targetClusterId).NamespaceId(namespaceId).DeploymentId(deploymentId)

	bkpModel, res, err := bkpClient.BackupServiceListBackupsExecute(listBkpRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}

	err = copier.Copy(&bkpResponse, bkpModel.Backups)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying the backup response: %v\n", err)
	}

	return bkpResponse, nil
}

// GetBackup will fetch backup for a given deployment
func (backup *PDS_API_V1) GetBackup(getBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("GetBackup is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}
