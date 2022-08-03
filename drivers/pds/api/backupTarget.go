package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// BackupTarget struct
type BackupTarget struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListBackupTarget func
func (backupTarget *BackupTarget) ListBackupTarget(tenantID string) ([]pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	backupTargetModels, res, err := backupTargetClient.ApiTenantsIdBackupTargetsGet(backupTarget.context, tenantID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupTargetsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModels.GetData(), err
}

// LisBackupsStateBelongToBackupTarget func
func (backupTarget *BackupTarget) LisBackupsStateBelongToBackupTarget(backuptargetID string) ([]pds.ModelsBackupTargetState, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdStatesGet(backupTarget.context, backuptargetID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdStatesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel.GetData(), err
}

// GetBackupTarget func
func (backupTarget *BackupTarget) GetBackupTarget(backupTargetID string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdGet(backupTarget.context, backupTargetID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err
}

// CreateBackupTarget func
func (backupTarget *BackupTarget) CreateBackupTarget(tenantID string, name string, backupCredentialsID string, bucket string, region string, backupType string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	createRequest := pds.ControllersCreateTenantBackupTarget{
		BackupCredentialsId: &backupCredentialsID,
		Bucket:              &bucket,
		Name:                &name,
		Region:              &region,
		Type:                &backupType,
	}
	backupTargetModel, res, err := backupTargetClient.ApiTenantsIdBackupTargetsPost(backupTarget.context, tenantID).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupTargetsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err

}

// UpdateBackupTarget func
func (backupTarget *BackupTarget) UpdateBackupTarget(backupTaregetID string, name string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	updateRequest := pds.ControllersUpdateBackupTargetRequest{
		Name: &name,
	}
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdPut(backupTarget.context, backupTaregetID).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err

}

// SyncToBackupLocation func
func (backupTarget *BackupTarget) SyncToBackupLocation(backupTaregetID string, name string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	updateRequest := pds.ControllersUpdateBackupTargetRequest{
		Name: &name,
	}
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdPut(backupTarget.context, backupTaregetID).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err
}

// DeleteBackupTarget func
func (backupTarget *BackupTarget) DeleteBackupTarget(backupTaregetID string) (*status.Response, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	res, err := backupTargetClient.ApiBackupTargetsIdDelete(backupTarget.context, backupTaregetID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
