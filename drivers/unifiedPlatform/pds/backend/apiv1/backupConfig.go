package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// getBackupConfigClient updates the header with bearer token and returns the new client
func (ds *PDSV2_API) getBackupConfigClient() (context.Context, *pdsv2.BackupConfigServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.ApiClientV2.BackupConfigServiceAPI

	return ctx, client, nil
}

// CreateBackupConfig will create backup config for a given deployment
func (ds *PDSV2_API) CreateBackupConfig(createBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := apiStructs.WorkFlowResponse{}

	_, backupClient, err := ds.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceCreateBackupConfigExecute(createBackupConfigRequest.BackupConfig.Create.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceCreateBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&backupConfigResponse, backupConfigModel)
	return &backupConfigResponse, err

}

// UpdateBackupConfig will update backup config for a given deployment
func (ds *PDSV2_API) UpdateBackupConfig(updateBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := apiStructs.WorkFlowResponse{}

	_, backupClient, err := ds.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceUpdateBackupConfigExecute(updateBackupConfigRequest.BackupConfig.Update.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceUpdateBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&backupConfigResponse, backupConfigModel)
	return &backupConfigResponse, err

}

// GetBackupConfig will fetch backup config for a given deployment
func (ds *PDSV2_API) GetBackupConfig(getBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := apiStructs.WorkFlowResponse{}

	_, backupClient, err := ds.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceGetBackupConfigExecute(getBackupConfigRequest.BackupConfig.Get.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceGetBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&backupConfigResponse, backupConfigModel)
	return &backupConfigResponse, err

}

// DeleteBackupConfig will delete backup config for a given deployment
func (ds *PDSV2_API) DeleteBackupConfig(deleteBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {

	_, backupClient, err := ds.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	_, res, err := backupClient.BackupConfigServiceDeleteBackupConfigExecute(deleteBackupConfigRequest.BackupConfig.Delete.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceDeleteBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil, err

}

// ListBackupConfig will list backup config for a given deployment
func (ds *PDSV2_API) ListBackupConfig(listBackupConfigRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := []apiStructs.WorkFlowResponse{}

	_, backupClient, err := ds.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceListBackupConfigsExecute(listBackupConfigRequest.BackupConfig.List.V1)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceListBackupConfigsExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&backupConfigResponse, backupConfigModel)
	return backupConfigResponse, err

}
