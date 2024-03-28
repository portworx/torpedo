package stworkflows

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
)

type WorkflowPDSBackup struct {
	WorkflowBackupConfig WorkflowPDSBackupConfig
}

// GetBackupIDByName returns the ID of given backup
func (backup WorkflowPDSBackup) GetBackupIDByName(name string, backupConfigName string, namespace string, deploymentName string) (string, error) {

	params := pdslibs.BackupParams{
		ProjectId:       backup.WorkflowBackupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId,
		DeploymentID:    backup.WorkflowBackupConfig.WorkflowDataService.Namespace.TargetCluster.ClusterUID,
		NamespaceId:     backup.WorkflowBackupConfig.WorkflowDataService.Namespace.Namespaces[namespace],
		TargetClusterId: backup.WorkflowBackupConfig.WorkflowDataService.Namespace.TargetCluster.ClusterUID,
		BackupConfigId:  *backup.WorkflowBackupConfig.Backups[backupConfigName].Meta.Uid,
	}

	allBackups, err := pdslibs.ListBackup(params)

	if err != nil {
		return "", err
	}

	for _, eachBackup := range allBackups.List.Backups {
		if *eachBackup.Meta.Name == name {
			return *eachBackup.Meta.Uid, nil
		}
	}

	return "", fmt.Errorf("[%s] - Backup not found ", name)
}

// GetBackupIDByName deletes the given backup
func (backup WorkflowPDSBackup) DeleteBackup(id string) error {
	err := pdslibs.DeleteBackupConfig(id)
	return err
}

// ListAllBackups returns the list of all backups
func (backup WorkflowPDSBackup) ListAllBackups(backupConfigName string, namespace string, deployment string) (*automationModels.PDSBackupResponse, error) {

	params := pdslibs.BackupParams{
		ProjectId:       backup.WorkflowBackupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId,
		DeploymentID:    backup.WorkflowBackupConfig.WorkflowDataService.Namespace.TargetCluster.ClusterUID,
		NamespaceId:     backup.WorkflowBackupConfig.WorkflowDataService.Namespace.Namespaces[namespace],
		TargetClusterId: backup.WorkflowBackupConfig.WorkflowDataService.Namespace.TargetCluster.ClusterUID,
		BackupConfigId:  *backup.WorkflowBackupConfig.Backups[backupConfigName].Meta.Uid,
	}

	list, err := pdslibs.ListBackup(params)

	if err != nil {
		return nil, err
	}

	return list, nil
}
