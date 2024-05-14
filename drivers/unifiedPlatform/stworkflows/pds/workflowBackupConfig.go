package pds

import (
	"fmt"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
	"strings"
	"time"
)

type WorkflowPDSBackupConfig struct {
	Backups                map[string]automationModels.V1BackupConfig
	WorkflowDataService    *WorkflowDataService
	SkipValidatation       map[string]bool
	WorkflowBackupLocation platform.WorkflowBackupLocation
}

const (
	ValidatePdsBackupConfig = "VALIDATE_PDS_BACKUP"
	RunDataBeforeBackup     = "RUN_DATA_BEFORE_BACKUP"
)

// CreateBackupConfig creates a backup config
func (backupConfig WorkflowPDSBackupConfig) CreateBackupConfig(name string, deploymentId string) (*automationModels.PDSBackupConfigResponse, error) {

	log.Infof("Backup [%s] started at [%s]", name, time.Now().Format("2006-01-02 15:04:05"))

	log.Infof("Backup name - [%s]", name)
	log.Infof("Deployment Name - [%s]", *backupConfig.WorkflowDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name)
	log.Infof("Delplyment UID - [%s]", deploymentId)
	log.Infof("Project Id - [%s]", backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId)
	log.Infof("Backup Location Id - [%s]", backupConfig.WorkflowBackupLocation.BkpLocation.BkpLocationId)

	if value, ok := backupConfig.SkipValidatation[RunDataBeforeBackup]; ok {
		if value == true {
			log.Infof("Skipping data insertion before backup")
		}
	} else {
		_, err := backupConfig.WorkflowDataService.RunDataServiceWorkloads(deploymentId)
		if err != nil {
			return nil, fmt.Errorf("unable to run workfload on data service. Error - [%s]", err.Error())
		}
	}

	createBackup, err := pdslibs.CreateBackupConfig(name,
		deploymentId,
		backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.ProjectId,
		backupConfig.WorkflowBackupLocation.BkpLocation.BkpLocationId)

	if err != nil {
		return nil, err
	}

	// TODO: Wait for backup to complete is to be implemented

	backupConfig.Backups[name] = createBackup.Create
	log.Infof("Backup config creates - Name - [%s] - ID - [%s]", *createBackup.Create.Meta.Name, *createBackup.Create.Meta.Uid)

	if value, ok := backupConfig.SkipValidatation[ValidatePdsBackupConfig]; ok {
		if value == true {
			log.Infof("Skipping Backup Validation")
		}
	} else {
		err = pdslibs.ValidateAdhocBackup(deploymentId, *createBackup.Create.Meta.Uid)
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

// Purge deletes all the backup config created during automation
func (backupConfig WorkflowPDSBackupConfig) Purge(ignoreError bool) error {

	log.Infof("Total number of backup configs found under [%s] are [%d]", backupConfig.WorkflowDataService.Namespace.TargetCluster.Project.Platform.TenantId, len(backupConfig.Backups))

	for _, eachBackupConfig := range backupConfig.Backups {
		log.Infof("Backup to be deleted - [%s]", *eachBackupConfig.Meta.Uid)
		err := pdslibs.DeleteBackupConfig(*eachBackupConfig.Meta.Uid)
		if err != nil {
			if ignoreError && !strings.Contains(err.Error(), "404 Not Found") {
				return err
			}
		}
		err = backupConfig.ValidateBackupConfigDeletion(*eachBackupConfig.Meta.Uid)
		if err != nil {
			return err
		}
		delete(backupConfig.Backups, *eachBackupConfig.Meta.Name)
		log.Infof("Backup config deleted - [%s]", *eachBackupConfig.Meta.Name)

	}

	return nil
}

func (backupConfig WorkflowPDSBackupConfig) ValidateBackupConfigDeletion(backupConfgId string) error {
	validateBackupDeletion := func() (interface{}, bool, error) {
		backupConfigDetails, err := pdslibs.GetBackupConfig(backupConfgId)
		if err == nil {
			return nil, true, fmt.Errorf("Backup Config [%s] is yet not deleted. Phase - [%v]", backupConfgId, *backupConfigDetails.Get.Status.Phase)
		} else {
			log.Infof("Backup [%s] is deleted successfully", backupConfgId)
			return nil, false, nil
		}
	}

	_, err := task.DoRetryWithTimeout(validateBackupDeletion, backupDeleteTimeOut, defaultRetryInterval)

	return err
}
