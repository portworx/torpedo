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
func (backup WorkflowPDSBackup) GetBackupIDByName(name string, backupConfigName string) (string, error) {

	allBackups, err := pdslibs.ListBackup(*backup.WorkflowBackupConfig.Backups[backupConfigName].Meta.Uid)

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
func (backup WorkflowPDSBackup) ListAllBackups(backupConfigName string) (*automationModels.PDSBackupResponse, error) {

	list, err := pdslibs.ListBackup(*backup.WorkflowBackupConfig.Backups[backupConfigName].Meta.Uid)

	if err != nil {
		return nil, err
	}

	return list, nil
}
