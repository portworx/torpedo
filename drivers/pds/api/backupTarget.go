package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type BackupTarget struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (backupTarget *BackupTarget) ListBackupTarget(tenantId string) ([]pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	backupTargetModels, res, err := backupTargetClient.ApiTenantsIdBackupTargetsGet(backupTarget.context, tenantId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupTargetsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModels.GetData(), err
}

func (backupTarget *BackupTarget) LisBackupsStateBelongToBackupTarget(backuptargetId string) ([]pds.ModelsBackupTargetState, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdStatesGet(backupTarget.context, backuptargetId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdStatesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel.GetData(), err
}

func (backupTarget *BackupTarget) GetBackupTarget(backupTargetId string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdGet(backupTarget.context, backupTargetId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err
}

func (backupTarget *BackupTarget) CreateBackupTarget(tenantId string, name string, backupCredentialsId string, bucket string, region string, backupType string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	createRequest := pds.ControllersCreateTenantBackupTarget{
		BackupCredentialsId: &backupCredentialsId,
		Bucket:              &bucket,
		Name:                &name,
		Region:              &region,
		Type:                &backupType,
	}
	backupTargetModel, res, err := backupTargetClient.ApiTenantsIdBackupTargetsPost(backupTarget.context, tenantId).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupTargetsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err

}
func (backupTarget *BackupTarget) UpdateBackupTarget(backupTaregetId string, name string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	updateRequest := pds.ControllersUpdateBackupTargetRequest{
		Name: &name,
	}
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdPut(backupTarget.context, backupTaregetId).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err

}
func (backupTarget *BackupTarget) SyncToBackupLocation(backupTaregetId string, name string) (*pds.ModelsBackupTarget, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	updateRequest := pds.ControllersUpdateBackupTargetRequest{
		Name: &name,
	}
	backupTargetModel, res, err := backupTargetClient.ApiBackupTargetsIdPut(backupTarget.context, backupTaregetId).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupTargetModel, err
}

func (backupTarget *BackupTarget) DeleteBackupTarget(backupTaregetId string) (*status.Response, error) {
	backupTargetClient := backupTarget.apiClient.BackupTargetsApi
	res, err := backupTargetClient.ApiBackupTargetsIdDelete(backupTarget.context, backupTaregetId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupTargetsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
