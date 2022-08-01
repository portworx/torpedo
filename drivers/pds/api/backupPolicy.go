package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type BackupPolicy struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (backupPolicy *BackupPolicy) ListBackupPolicy(tenantId string) ([]pds.ModelsBackupPolicy, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	backupModels, res, err := backupClient.ApiTenantsIdBackupPoliciesGet(backupPolicy.context, tenantId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupPoliciesGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}

func (backupPolicy *BackupPolicy) GetBackupPolicy(backupCredId string) (*pds.ModelsBackupPolicy, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	backupPolicyModel, res, err := backupClient.ApiBackupPoliciesIdGet(backupPolicy.context, backupCredId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupPoliciesIdGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupPolicyModel, err
}

func (backupPolicy *BackupPolicy) CreateBackupPolicy(tenantId string, name string, retentionCount int32, scheduleCronExpression string, backupType string) (*pds.ModelsBackupPolicy, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	modelBackupSchedule := []pds.ModelsBackupSchedule{{
		RetentionCount: &retentionCount,
		Schedule:       &scheduleCronExpression,
		Type:           &backupType,
	}}
	createRequest := pds.ControllersCreateBackupPolicyRequest{
		Schedules: modelBackupSchedule,
		Name:      &name,
	}
	backupPolicyModel, res, err := backupClient.ApiTenantsIdBackupPoliciesPost(backupPolicy.context, tenantId).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupPoliciesPost``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupPolicyModel, err

}
func (backupPolicy *BackupPolicy) UpdateBackupPolicy(backupCredsId string, name string, retentionCount int32, scheduleCronExpression string, backupType string) (*pds.ModelsBackupPolicy, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	modelBackupSchedule := []pds.ModelsBackupSchedule{{
		RetentionCount: &retentionCount,
		Schedule:       &scheduleCronExpression,
		Type:           &backupType,
	}}
	updateRequest := pds.ControllersUpdateBackupPolicyRequest{
		Schedules: modelBackupSchedule,
		Name:      &name,
	}
	backupPolicyModel, res, err := backupClient.ApiBackupPoliciesIdPut(backupPolicy.context, backupCredsId).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupPoliciesIdPut``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupPolicyModel, err

}

func (backupPolicy *BackupPolicy) DeleteBackupPolicy(backupCredsId string) (*status.Response, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	res, err := backupClient.ApiBackupPoliciesIdDelete(backupPolicy.context, backupCredsId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupPoliciesIdDelete``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
