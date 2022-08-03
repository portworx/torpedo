package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// BackupPolicy struct
type BackupPolicy struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListBackupPolicy func
func (backupPolicy *BackupPolicy) ListBackupPolicy(tenantID string) ([]pds.ModelsBackupPolicy, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	backupModels, res, err := backupClient.ApiTenantsIdBackupPoliciesGet(backupPolicy.context, tenantID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupPoliciesGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}

// GetBackupPolicy func
func (backupPolicy *BackupPolicy) GetBackupPolicy(backupCredID string) (*pds.ModelsBackupPolicy, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	backupPolicyModel, res, err := backupClient.ApiBackupPoliciesIdGet(backupPolicy.context, backupCredID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupPoliciesIdGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupPolicyModel, err
}

// CreateBackupPolicy func
func (backupPolicy *BackupPolicy) CreateBackupPolicy(tenantID string, name string, retentionCount int32, scheduleCronExpression string, backupType string) (*pds.ModelsBackupPolicy, error) {
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
	backupPolicyModel, res, err := backupClient.ApiTenantsIdBackupPoliciesPost(backupPolicy.context, tenantID).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupPoliciesPost``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupPolicyModel, err

}

// UpdateBackupPolicy func
func (backupPolicy *BackupPolicy) UpdateBackupPolicy(backupCredsID string, name string, retentionCount int32, scheduleCronExpression string, backupType string) (*pds.ModelsBackupPolicy, error) {
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
	backupPolicyModel, res, err := backupClient.ApiBackupPoliciesIdPut(backupPolicy.context, backupCredsID).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupPoliciesIdPut``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
	}
	return backupPolicyModel, err

}

// DeleteBackupPolicy func
func (backupPolicy *BackupPolicy) DeleteBackupPolicy(backupCredsID string) (*status.Response, error) {
	backupClient := backupPolicy.apiClient.BackupPoliciesApi
	res, err := backupClient.ApiBackupPoliciesIdDelete(backupPolicy.context, backupCredsID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupPoliciesIdDelete``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
