package pds

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"time"

	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSBackup struct {
	WorkflowDataService WorkflowDataService
}

const (
	defaultRetryInterval = 10 * time.Second
	backupTimeOut        = 20 * time.Minute
	backupDeleteTimeOut  = 5 * time.Minute
)

// GetBackupIDByName returns the ID of given backup
func (backup WorkflowPDSBackup) GetLatestBackup(deploymentName string) (automationModels.V1Backup, error) {

	var latestBackup automationModels.V1Backup

	allBackups, err := pdslibs.ListBackup(backup.WorkflowDataService.DataServiceDeployment[deploymentName])

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
		if *backupModel.Get.Status.Phase != stworkflows.COMPLETED {
			return nil, true, fmt.Errorf("Backup is not completed yet, Phase - [%s]", *backupModel.Get.Status.Phase)
		} else {
			log.Infof("Backup completed successfully - [%s]", *backupModel.Get.Meta.Name)
			log.Infof("Backup Status - [%s]", *backupModel.Get.Status.CloudSnapId)
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
func (backup WorkflowPDSBackup) ListAllBackups(deploymentName string) ([]automationModels.V1Backup, error) {

	allBackups := make([]automationModels.V1Backup, 0)

	response, err := pdslibs.ListBackup(backup.WorkflowDataService.DataServiceDeployment[deploymentName])

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
func (backup WorkflowPDSBackup) Purge(deploymentName string) error {

	allBackups, err := backup.ListAllBackups(deploymentName)
	if err != nil {
		return err
	}

	log.Infof("Total number of backups found for [%s] are [%d]", deploymentName, len(allBackups))

	for _, eachBackup := range allBackups {
		err := backup.DeleteBackup(*eachBackup.Meta.Uid)
		if err != nil {
			return err
		}
		err = backup.ValidateBackupDeletion(*eachBackup.Meta.Uid)
		if err != nil {
			return fmt.Errorf("Backup deleted but validation failed. Error - [%s]", err.Error())
		}
	}

	return nil
}
