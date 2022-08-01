package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type BackupJob struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (backupJob *BackupJob) ListBackupJobs(backupId string) ([]pds.ControllersBackupJobStatus, error) {
	backupJobClient := backupJob.apiClient.BackupJobsApi
	backupJobModels, res, err := backupJobClient.ApiBackupsIdJobsGet(backupJob.context, backupId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdJobsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupJobModels.GetData(), err
}

func (backupJob *BackupJob) GetBackupJob(backupJobId string) (*pds.ModelsBackupJob, error) {
	backupJobClient := backupJob.apiClient.BackupJobsApi
	backupJobModel, res, err := backupJobClient.ApiBackupJobsIdGet(backupJob.context, backupJobId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupJobsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupJobModel, err
}
