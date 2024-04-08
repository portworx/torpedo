package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// DeleteBackup deletes backup config of the deployment
func DeleteBackup(backupId string) error {

	deleteBackupRequest := automationModels.PDSBackupRequest{
		Delete: automationModels.PDSDeleteBackup{
			Id: backupId,
		},
	}

	err := v2Components.PDS.DeleteBackup(&deleteBackupRequest)
	if err != nil {
		return err
	}
	return err
}

// GetBackup fetches backup config for the deployment
func GetBackup(backupId string) (*automationModels.PDSBackupResponse, error) {

	getBackupRequest := automationModels.PDSBackupRequest{
		Get: automationModels.PDSGetBackup{
			Id: backupId,
		},
	}

	backupResponse, err := v2Components.PDS.GetBackup(&getBackupRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err

}

// ListBackup lists backup config for the deployment
func ListBackup(backupConfigId string) (*automationModels.PDSBackupResponse, error) {

	listBackup := automationModels.PDSBackupRequest{
		List: automationModels.PDSListBackup{
			BackupConfigId: backupConfigId,
		},
	}

	backupResponse, err := v2Components.PDS.ListBackup(&listBackup)
	if err != nil {
		return nil, err
	}

	totalRecords := *backupResponse.List.Pagination.TotalRecords
	log.Infof("Total backup of  [%s] = [%s]", backupConfigId, totalRecords)

	listBackup = automationModels.PDSBackupRequest{
		List: automationModels.PDSListBackup{
			BackupConfigId:       backupConfigId,
			PaginationPageNumber: DEFAULT_PAGE_NUMBER,
			PaginationPageSize:   totalRecords,
			SortSortBy:           DEFAULT_SORT_BY,
			SortSortOrder:        DEFAULT_SORT_ORDER,
		},
	}

	backupResponse, err = v2Components.PDS.ListBackup(&listBackup)
	if err != nil {
		return nil, err
	}

	return backupResponse, err
}
