package pds

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
)

type WorkflowPDSBackup struct {
	WorkflowDataService WorkflowDataService
}

// GetBackupIDByName returns the ID of given backup
func (backup WorkflowPDSBackup) GetLatestBackup(deploymentName string) (automationModels.V1Backup, error) {

	var latestBackup automationModels.V1Backup

	allBackups, err := pdslibs.ListBackup(backup.WorkflowDataService.DataServiceDeployment[deploymentName])

	if err != nil {
		return latestBackup, err
	}

	latestBackup = allBackups.List.Backups[0]

	return latestBackup, nil
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
