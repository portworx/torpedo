package pds

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

type WorkflowPDSBackup struct {
	WorkflowDataService WorkflowDataService
}

const (
	DefaultRetryInterval = 10 * time.Second
	backupTimeOut        = 10 * time.Minute
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

	err := wait.Poll(DefaultRetryInterval, backupTimeOut, func() (bool, error) {
		backupModel, err := pdslibs.GetBackup(backupId)
		if *backupModel.Get.Status.Phase != stworkflows.COMPLETED {
			return false, fmt.Errorf("Backup is not completed yet, Phase - [%s]", err.Error())
		}
		log.Infof("Backup completed successfully - [%s]", *backupModel.Get.Meta.Name)
		return true, nil
	})

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
