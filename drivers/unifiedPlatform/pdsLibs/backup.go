package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

type WorkflowBackup struct {
	ProjectId        string
	DeploymentID     string
	NamespaceId      string
	TargetClusterId  string
	BackupConfigId   string
}

// DeleteBackup deletes backup config of the deployment
func DeleteBackup(backup WorkflowBackup) ( error) {

	deleteBackupRequest := automationModels.PDSBackupRequest{}

	deleteBackupRequest.Delete.Id = backup.BackupConfigId

	err := v2Components.PDS.DeleteBackup(&deleteBackupRequest)
	if err != nil {
		return err
	}
	return err
}

// GetBackup fetches backup config for the deployment
func GetBackup(backup WorkflowBackup) (*automationModels.PDSBackupResponse, error) {

	getBackupRequest := automationModels.PDSBackupRequest{}

	getBackupRequest.Get.Id = backup.BackupConfigId

	backupResponse, err := v2Components.PDS.GetBackup(&getBackupRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err

}

// ListBackup lists backup config for the deployment
func ListBackup(backup WorkflowBackup) ([]automationModels.PDSBackupResponse, error) {

	listBackup := automationModels.PDSBackupRequest{}

	listBackup.List.TargetClusterId = backup.TargetClusterId
	listBackup.List.NamespaceId = backup.NamespaceId
	listBackup.List.DeploymentId = backup.DeploymentID
	listBackup.List.BackupConfigId = backup.BackupConfigId


	backupResponse, err := v2Components.PDS.ListBackup(&listBackup)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
