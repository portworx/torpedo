package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSBackupConfig struct {
	Backups             map[string]automationModels.V1BackupConfig
	WorkflowDataService WorkflowDataService
	SkipValidatation    map[string]bool
}

const (
	ValidatePdsBackupConfig = "VALIDATE_PDS_BACKUP"
)

// CreateBackupConfig creates a backup config
func (backupConfig WorkflowPDSBackupConfig) CreateBackupConfig(name string, deploymentName string) (*automationModels.PDSBackupConfigResponse, error) {
	createBackup, err := pdslibs.CreateBackupConfig(name,
		backupConfig.WorkflowDataService.DataServiceDeployment[deploymentName],
		backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId)

	if err != nil {
		return nil, err
	}

	backupConfig.Backups[name] = createBackup.Create

	if value, ok := backupConfig.SkipValidatation[ValidatePdsBackupConfig]; ok {
		if value == true {
			log.Infof("Skipping Backup Validation")
		}
	} else {
		var bkp pdslibs.WorkflowBackup
		bkp.BackupConfigId = *createBackup.Create.Meta.Uid
		bkp.TargetClusterId = backupConfig.WorkflowDataService.Namespace.TargetCluster.ClusterUID
		bkp.NamespaceId = backupConfig.WorkflowDataService.Namespace.Namespaces[backupConfig.WorkflowDataService.NamespaceName]
		bkp.DeploymentID = backupConfig.WorkflowDataService.DataServiceDeployment[deploymentName]
		err = pdslibs.ValidateAdhocBackup(bkp)
		if err != nil {
			return nil, err
		}
	}

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
