package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Backup struct
type Backup struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListBackup func
func (backup *Backup) ListBackup(deploymentID string) ([]pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backupModels, res, err := backupClient.ApiDeploymentsIdBackupsGet(backup.context, deploymentID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdBackupsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}

// ListBackupsBelongToTarget func
func (backup *Backup) ListBackupsBelongToTarget(backupTargetID string) ([]pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backupModels, res, err := backupClient.ApiBackupTargetsIdBackupsGet(backup.context, backupTargetID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdBackupsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}

// GetBackup func
func (backup *Backup) GetBackup(backupID string) (*pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backuptModel, res, err := backupClient.ApiBackupsIdGet(backup.context, backupID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backuptModel, err
}

// CreateBackup func
func (backup *Backup) CreateBackup(deploymentID string, backupTargetID string, jobHistoryLimit int32, isAdhoc bool) (*pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	backupType := "adhoc"
	if !isAdhoc {
		backupType = "scheduled"
	}
	backupLevel := "snapshot"
	createRequest := pds.ControllersCreateDeploymentBackup{
		BackupLevel:     &backupLevel,
		BackupTargetId:  &backupTargetID,
		BackupType:      &backupType,
		JobHistoryLimit: &jobHistoryLimit,
	}
	backuptModel, res, err := backupClient.ApiDeploymentsIdBackupsPost(backup.context, deploymentID).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdBackupsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backuptModel, err

}

// UpdateBackup func
func (backup *Backup) UpdateBackup(backupID string, jobHistoryLimit int32) (*pds.ModelsBackup, error) {
	backupClient := backup.apiClient.BackupsApi
	updateRequest := pds.ControllersUpdateBackupRequest{
		JobHistoryLimit: &jobHistoryLimit,
	}
	backupTargetModel, res, err := backupClient.ApiBackupsIdPut(backup.context, backupID).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err

}

// DeleteBackupJobs func
func (backup *Backup) DeleteBackupJobs(backupID string, jobName string) (*status.Response, error) {
	backupClient := backup.apiClient.BackupsApi
	res, err := backupClient.ApiBackupsIdJobsNameDelete(backup.context, backupID, jobName).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdJobsNameDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}

// DeleteBackup func
func (backup *Backup) DeleteBackup(backupID string) (*status.Response, error) {
	backupClient := backup.apiClient.BackupsApi
	res, err := backupClient.ApiBackupsIdDelete(backup.context, backupID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
