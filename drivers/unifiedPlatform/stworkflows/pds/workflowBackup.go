package pds

import (
	"fmt"
	"time"

	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSBackup struct {
	WorkflowDataService  *WorkflowDataService
	Backups              map[string]*BackupDetails
	Md5HashesForBackusp  map[string]string
	WorkflowBackupConfig *WorkflowPDSBackupConfig
}

type BackupDetails struct {
	Backup         *automationModels.V1Backup
	BackupConfigId string
	Md5Hash        string
	TableName      string
}

const (
	defaultRetryInterval = 10 * time.Second
	backupTimeOut        = 20 * time.Minute
	backupDeleteTimeOut  = 5 * time.Minute
)

// GetBackupIDByName returns the ID of given backup
func (backup WorkflowPDSBackup) GetLatestBackup(deploymentId string) (automationModels.V1Backup, error) {

	var latestBackup automationModels.V1Backup

	allBackups, err := pdslibs.ListBackup(deploymentId)

	if err != nil {
		return latestBackup, err
	}

	if len(allBackups.List.Backups) > 0 {
		latestBackup = allBackups.List.Backups[0]
	} else {
		return latestBackup, fmt.Errorf("No backups found for backup config")
	}

	return latestBackup, nil
}

func (backup WorkflowPDSBackup) WaitForBackupToComplete(backupId string) error {

	waitforBackupToComplete := func() (interface{}, bool, error) {
		backupModel, err := pdslibs.GetBackup(backupId)
		if err != nil {
			return nil, false, fmt.Errorf("Some error occurred while polling for backup. Error - [%s]", err.Error())
		}
		if *backupModel.Get.Status.Phase == stworkflows.FAILED {
			return nil, false, fmt.Errorf("Backup Status - [%s]", *backupModel.Get.Status.Phase)
		} else if *backupModel.Get.Status.Phase != stworkflows.COMPLETED {
			return nil, true, fmt.Errorf("Backup is not completed yet, Phase - [%s]", *backupModel.Get.Status.Phase)
		} else {
			log.Infof("Backup completed successfully - [%s]", *backupModel.Get.Meta.Name)
			log.Infof("Backup Status - [%s]", *backupModel.Get.Status.CloudSnapId)
			backup.Backups[*backupModel.Get.Meta.Uid] = &BackupDetails{
				Backup:         &backupModel.Get,
				BackupConfigId: *backupModel.Get.Meta.ParentReference.Uid,
				Md5Hash:        backup.WorkflowBackupConfig.BackupConfigs[*backupModel.Get.Meta.ParentReference.Uid].Md5,
				TableName:      backup.WorkflowBackupConfig.BackupConfigs[*backupModel.Get.Meta.ParentReference.Uid].ChkSumMap[*backupModel.Get.Meta.ParentReference.Uid],
			}
			return nil, false, nil
		}
	}

	_, err := task.DoRetryWithTimeout(waitforBackupToComplete, backupTimeOut, defaultRetryInterval)

	return err
}

// GetBackupIDByName deletes the given backup
func (backup WorkflowPDSBackup) DeleteBackup(id string) error {
	err := pdslibs.DeleteBackup(id)
	return err
}

// ListAllBackups lists all backups
func (backup WorkflowPDSBackup) ListAllBackups(deploymentId string) ([]automationModels.V1Backup, error) {

	allBackups := make([]automationModels.V1Backup, 0)

	response, err := pdslibs.ListBackup(deploymentId)

	if err != nil {
		return allBackups, err
	}

	return response.List.Backups, nil
}

func (backup WorkflowPDSBackup) ValidateBackupDeletion(id string) error {

	validateBackupDeletion := func() (interface{}, bool, error) {
		backupDetails, err := pdslibs.GetBackup(id)
		if err == nil {
			return nil, true, fmt.Errorf("Backup [%s] is yet not deleted. Phase - [%v]", id, *backupDetails.Get.Status.Phase)
		} else {
			log.Infof("Backup [%s] is deleted successfully", id)
			return nil, false, nil
		}
	}

	_, err := task.DoRetryWithTimeout(validateBackupDeletion, backupDeleteTimeOut, defaultRetryInterval)

	return err
}

// Purge deletes all backups for a given deployment
func (backup WorkflowPDSBackup) Purge() error {

	log.Infof("Total number of backups found - [%d]", len(backup.Backups))

	for eachBackup, backupDetails := range backup.Backups {
		log.InfoD("Deleting [%s]", *backupDetails.Backup.Meta.Name)
		err := backup.DeleteBackup(eachBackup)
		if err != nil {
			return err
		}
		err = backup.ValidateBackupDeletion(eachBackup)
		if err != nil {
			return fmt.Errorf("Backup deleted but validation failed. Error - [%s]", err.Error())
		}
	}

	return nil
}
