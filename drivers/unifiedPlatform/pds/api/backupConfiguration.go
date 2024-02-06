package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// BackupConfigV2 struct
type BackupConfigV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (bckpConfig *BackupConfigV2) GetClient() (context.Context, *pdsv2.BackupConfigServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	bckpConfig.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	bckpConfig.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = bckpConfig.AccountID
	client := bckpConfig.ApiClientV2.BackupConfigServiceAPI
	return ctx, client, nil
}

// ListBackupConfigurations return pds backup config models.
func (bckpConfig *BackupConfigV2) ListBackupConfigurations() ([]pdsv2.V1BackupConfig, error) {
	ctx, backupClient, err := bckpConfig.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModels, res, err := backupClient.BackupConfigServiceListBackupConfigs(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceListBackupConfigs`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModels.BackupConfigs, err
}

// GetBackupConfigurations return pds backup config model.
func (bckpConfig *BackupConfigV2) GetBackupConfigurations(backupConfigID string) (*pdsv2.V1BackupConfig, error) {
	ctx, backupClient, err := bckpConfig.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModel, res, err := backupClient.BackupConfigServiceGetBackupConfig(ctx, backupConfigID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceGetBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModel, err
}

// CreateBackupConfigurations create backup config models.
func (bckpConfig *BackupConfigV2) CreateBackupConfigurations(projectId string) (*pdsv2.V1BackupConfig, error) {
	ctx, backupClient, err := bckpConfig.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModel, res, err := backupClient.BackupConfigServiceCreateBackupConfig(ctx, projectId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceCreateBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModel, err
}

// UpdateBackupConfigurations return updated backup config model.
func (bckpConfig *BackupConfigV2) UpdateBackupConfigurations(projectId string) (*pdsv2.V1BackupConfig, error) {
	ctx, backupClient, err := bckpConfig.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupModel, res, err := backupClient.BackupConfigServiceUpdateBackupConfig(ctx, projectId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceUpdateBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return backupModel, err
}

// DeleteBackupConfigurations delete the backup and return the status.
func (bckpConfig *BackupConfigV2) DeleteBackupConfigurations(backupConfigID string) (*status.Response, error) {
	ctx, backupClient, err := bckpConfig.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := backupClient.BackupConfigServiceDeleteBackupConfig(ctx, backupConfigID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceDeleteBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
