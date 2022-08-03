package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// BackupJob struct
type BackupJob struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListBackupJobs func
func (backupJob *BackupJob) ListBackupJobs(backupID string) ([]pds.ControllersBackupJobStatus, error) {
	backupJobClient := backupJob.apiClient.BackupJobsApi
	backupJobModels, res, err := backupJobClient.ApiBackupsIdJobsGet(backupJob.context, backupID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdJobsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupJobModels.GetData(), err
}

// GetBackupJob func
func (backupJob *BackupJob) GetBackupJob(backupJobID string) (*pds.ModelsBackupJob, error) {
	backupJobClient := backupJob.apiClient.BackupJobsApi
	backupJobModel, res, err := backupJobClient.ApiBackupJobsIdGet(backupJob.context, backupJobID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupJobsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupJobModel, err
}
