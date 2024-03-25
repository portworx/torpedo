package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

// CreateBackupConfig created backup config for the deployment
func CreateBackupConfig(name string, deploymentId string, projectId string) (*automationModels.PDSBackupConfigResponse, error) {

	createBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	createBackupConfigRequest.Create.BackupConfig = &automationModels.V1BackupConfig{
		Meta: &automationModels.Meta{
			Uid: &name,
		},
	}
	createBackupConfigRequest.Create.DeploymentId = deploymentId
	createBackupConfigRequest.Create.ProjectId = projectId

	backupResponse, err := v2Components.PDS.CreateBackupConfig(&createBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// UpdateBackupConfig updates backup config of the deployment
func UpdateBackupConfig(id string, labels map[string]string, annotations map[string]string) (*automationModels.PDSBackupConfigResponse, error) {

	updateBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	updateBackupConfigRequest.Update.ID = id
	updateBackupConfigRequest.Update.Labels = &labels
	updateBackupConfigRequest.Update.Annotations = &annotations

	backupResponse, err := v2Components.PDS.UpdateBackupConfig(&updateBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// DeleteBackupConfig deletes backup config of the deployment
func DeleteBackupConfig(id string) error {

	deleteBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	deleteBackupConfigRequest.Delete.Id = id

	err := v2Components.PDS.DeleteBackupConfig(&deleteBackupConfigRequest)
	if err != nil {
		return err
	}
	return nil
}

// GetBackupConfig fetches backup config for the deployment
func GetBackupConfig(id string) (*automationModels.PDSBackupConfigResponse, error) {

	getBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	getBackupConfigRequest.Get.Id = id

	backupResponse, err := v2Components.PDS.GetBackupConfig(&getBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// ListBackupConfig lists backup config for the deployment
func ListBackupConfig(accountId string, tenantId string, projectId string, targetClusterId string, namespaceId string, deploymentId string) (*automationModels.PDSBackupConfigResponse, error) {

	listBackupConfig := automationModels.PDSBackupConfigRequest{}

	listBackupConfig.List.AccountId = &accountId
	listBackupConfig.List.TenantId = &tenantId
	listBackupConfig.List.ProjectId = &projectId
	listBackupConfig.List.TargetClusterId = &targetClusterId
	listBackupConfig.List.NamespaceId = &namespaceId
	listBackupConfig.List.DeploymentId = &deploymentId

	backupResponse, err := v2Components.PDS.ListBackupConfig(&listBackupConfig)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
