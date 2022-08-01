package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type Backup struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (backup *Backup) ListBackup(deploymentId string) ([]pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backupModels, res, err := backupClient.ApiDeploymentsIdBackupsGet(backup.context, deploymentId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdBackupsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}
func (backup *Backup) ListBackupsBelongToTarget(backupTargetId string) ([]pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backupModels, res, err := backupClient.ApiBackupTargetsIdBackupsGet(backup.context, backupTargetId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdBackupsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}

func (backup *Backup) GetBackup(backupId string) (*pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backuptModel, res, err := backupClient.ApiBackupsIdGet(backup.context, backupId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backuptModel, err
}

func (backup *Backup) CreateBackup(deploymentId string, backupTargetId string, jobHistoryLimit int32, isAdhoc bool) (*pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backupType := "adhoc"
	if !isAdhoc {
		backupType = "scheduled"
	}
	backupLevel := "snapshot"
	createRequest := pds.ControllersCreateDeploymentBackup{
		BackupLevel:     &backupLevel,
		BackupTargetId:  &backupTargetId,
		BackupType:      &backupType,
		JobHistoryLimit: &jobHistoryLimit,
	}
	backuptModel, res, err := backupClient.ApiDeploymentsIdBackupsPost(backup.context, deploymentId).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdBackupsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backuptModel, err

}

func (backup *Backup) UpdateBackup(backupId string, jobHistoryLimit int32) (*pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	updateRequest := pds.ControllersUpdateBackupRequest{
		JobHistoryLimit: &jobHistoryLimit,
	}
	backupTargetModel, res, err := backupClient.ApiBackupsIdPut(backup.context, backupId).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err

}
func (backup *Backup) DeleteBackupJobs(backupId string, jobName string) (*status.Response, error) {
	backupClient := backup.apiClient.BackupsApi
	res, err := backupClient.ApiBackupsIdJobsNameDelete(backup.context, backupId, jobName).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdJobsNameDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}

func (backup *Backup) DeleteBackup(backupId string) (*status.Response, error) {
	backupClient := backup.apiClient.BackupsApi
	res, err := backupClient.ApiBackupsIdDelete(backup.context, backupId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
