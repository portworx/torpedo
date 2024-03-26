package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
)

type WorkflowPDSBackupConfig struct {
	Backups             map[string]automationModels.V1BackupConfig
	WorkflowDataService WorkflowDataService
}

// CreateBackupConfig creates a backup config
func (backupConfig WorkflowPDSBackupConfig) CreateBackupConfig(name string, dataserviceName string) (*automationModels.PDSBackupConfigResponse, error) {
	createBackup, err := pdslibs.CreateBackupConfig(name,
		backupConfig.WorkflowDataService.DataServiceDeployment[dataserviceName],
		backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId)

	if err != nil {
		return nil, err
	}

	// TODO: Wait for backup to complete is to be implemented

	backupConfig.Backups[name] = createBackup.Create

	return createBackup, nil
}

// DeleteBackupConfig deletes a backup config
func (backupConfig WorkflowPDSBackupConfig) DeleteBackupConfig(name string) error {
	err := pdslibs.DeleteBackupConfig(*backupConfig.Backups[name].Meta.Uid)
	return err
}

// ListBackupConfig lists all backup config
func (backupConfig WorkflowPDSBackupConfig) ListBackupConfig(namespceName string, deploymentName string) (*automationModels.PDSBackupConfigResponse, error) {
	listBackups, err := pdslibs.ListBackupConfig(
		backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.Platform.AdminAccountId,
		backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.Platform.TenantId,
		backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId,
		backupConfig.WorkflowDataService.Namespace.TargetCluster.ClusterUID,
		backupConfig.WorkflowDataService.Namespace.Namespaces[namespceName],
		backupConfig.WorkflowDataService.DataServiceDeployment[deploymentName],
	)
	if err != nil {
		return nil, err
	}

	return listBackups, err
}
