package pds

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSBackupConfig struct {
	Backups                map[string]automationModels.V1BackupConfig
	WorkflowDataService    WorkflowDataService
	SkipValidatation       map[string]bool
	WorkflowBackupLocation platform.WorkflowBackupLocation
}

const (
	ValidatePdsBackupConfig = "VALIDATE_PDS_BACKUP"
)

// CreateBackupConfig creates a backup config
func (backupConfig WorkflowPDSBackupConfig) CreateBackupConfig(name string, deploymentName string) (*automationModels.PDSBackupConfigResponse, error) {

	log.Infof("Backup name - [%s]", name)
	log.Infof("Delplyment UID - [%s]", backupConfig.WorkflowDataService.DataServiceDeployment[deploymentName])
	log.Infof("Project Id - [%s]", backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId)
	log.Infof("Backup Location Id - [%s]", backupConfig.WorkflowBackupLocation.BkpLocation.BkpLocationId)

	createBackup, err := pdslibs.CreateBackupConfig(name,
		backupConfig.WorkflowDataService.DataServiceDeployment[deploymentName],
		backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId,
		backupConfig.WorkflowBackupLocation.BkpLocation.BkpLocationId)

	if err != nil {
		return nil, err
	}

	// TODO: Wait for backup to complete is to be implemented

	backupConfig.Backups[name] = createBackup.Create

	if value, ok := backupConfig.SkipValidatation[ValidatePdsBackupConfig]; ok {
		if value == true {
			log.Infof("Skipping Backup Validation")
		}
	} else {
		err = pdslibs.ValidateAdhocBackup(backupConfig.WorkflowDataService.DataServiceDeployment[deploymentName], *createBackup.Create.Meta.Uid)
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
func (backupConfig WorkflowPDSBackupConfig) ListBackupConfig(tenantId string) (*automationModels.PDSBackupConfigResponse, error) {
	listBackups, err := pdslibs.ListBackupConfig(tenantId)
	if err != nil {
		return nil, err
	}

	return listBackups, err
}
