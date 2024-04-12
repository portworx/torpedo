package pds

import (
	"fmt"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	"time"
)

type WorkflowPDSBackup struct {
	WorkflowDataService WorkflowDataService
}

const (
	defaultRetryInterval = 10 * time.Second
	backupTimeOut        = 20 * time.Minute
)

// GetBackupIDByName returns the ID of given backup
func (backup WorkflowPDSBackup) GetLatestBackup(deploymentName string) (automationModels.V1Backup, error) {

	var latestBackup automationModels.V1Backup

	log.Infof("All deployments - [%+v]", backup.WorkflowDataService.DataServiceDeployment)

	allBackups, err := pdslibs.ListBackup(backup.WorkflowDataService.DataServiceDeployment[deploymentName])

	if err != nil {
		return latestBackup, err
	}

	latestBackup = allBackups.List.Backups[0]

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
			return nil, false, nil
		}
	}

	_, err := task.DoRetryWithTimeout(waitforBackupToComplete, backupTimeOut, defaultRetryInterval)

	return err

}

// GetBackupIDByName deletes the given backup
func (backup WorkflowPDSBackup) DeleteBackup(id string) error {
	err := pdslibs.DeleteBackupConfig(id)
	return err
}

func (backup WorkflowPDSBackup) ListAllBackups(deploymentName string) ([]automationModels.V1Backup, error) {

	allBackups := make([]automationModels.V1Backup, 0)

	response, err := pdslibs.ListBackup(backup.WorkflowDataService.DataServiceDeployment[deploymentName])

	if err != nil {
		return allBackups, err
	}

	return response.List.Backups, nil
}
