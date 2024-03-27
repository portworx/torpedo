package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

// DeleteBackup deletes backup config of the deployment
func DeleteBackup(id string) error {

	deleteBackupRequest := automationModels.PDSBackupRequest{}

	deleteBackupRequest.Delete.Id = id

	err := v2Components.PDS.DeleteBackup(&deleteBackupRequest)
	if err != nil {
		return err
	}
	return err
}

// GetBackup fetches backup config for the deployment
func GetBackup(id string) (*automationModels.PDSBackupResponse, error) {

	getBackupRequest := automationModels.PDSBackupRequest{}

	getBackupRequest.Get.Id = id

	backupResponse, err := v2Components.PDS.GetBackup(&getBackupRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err

}

// ListBackup lists backup config for the deployment
func ListBackups(backupConfigId string, targetClusterId string, namespaceId string, deploymentId string) (*automationModels.PDSBackupResponse, error) {

	listBackup := automationModels.PDSBackupRequest{}

	listBackup.List.TargetClusterId = targetClusterId
	listBackup.List.NamespaceId = namespaceId
	listBackup.List.DeploymentId = deploymentId
	listBackup.List.BackupConfigId = backupConfigId

	backupResponse, err := v2Components.PDS.ListBackup(&listBackup)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
