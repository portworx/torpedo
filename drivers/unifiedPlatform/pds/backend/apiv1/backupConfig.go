package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	status "net/http"
)

// getBackupConfigClient updates the header with bearer token and returns the new client
func (backupConf *PDSV2_API) getBackupConfigClient() (context.Context, *pdsv2.BackupConfigServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	backupConf.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	backupConf.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = backupConf.AccountID
	client := backupConf.ApiClientV2.BackupConfigServiceAPI

	return ctx, client, nil
}

// CreateBackupConfig will create backup config for a given deployment
func (backupConf *PDSV2_API) CreateBackupConfig(createBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := apiStructs.WorkFlowResponse{}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigCreateRequest := backupClient.BackupConfigServiceCreateBackupConfig(ctx, createBackupConfigRequest.BackupConfig.V1.Create.ProjectId)

	err = utilities.CopyStruct(backupConfigCreateRequest, createBackupConfigRequest.BackupConfig.V1.Create)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceCreateBackupConfigExecute(backupConfigCreateRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceCreateBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&backupConfigResponse, backupConfigModel)
	return &backupConfigResponse, err

}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PDSV2_API) UpdateBackupConfig(updateBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := apiStructs.WorkFlowResponse{}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigUpdateRequest := backupClient.BackupConfigServiceUpdateBackupConfig(ctx, updateBackupConfigRequest.BackupConfig.V1.Update.BackupConfigMetaUid)
	err = utilities.CopyStruct(backupConfigUpdateRequest, updateBackupConfigRequest.BackupConfig.V1.Update)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceUpdateBackupConfigExecute(backupConfigUpdateRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceUpdateBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(&backupConfigResponse, backupConfigModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return &backupConfigResponse, err

}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PDSV2_API) GetBackupConfig(getBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := apiStructs.WorkFlowResponse{}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigGetRequest := backupClient.BackupConfigServiceGetBackupConfig(ctx, getBackupConfigRequest.BackupConfig.V1.Get.Id)
	err = utilities.CopyStruct(backupConfigGetRequest, getBackupConfigRequest.BackupConfig.V1.Get)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceGetBackupConfigExecute(backupConfigGetRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceGetBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&backupConfigResponse, backupConfigModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return &backupConfigResponse, err

}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PDSV2_API) DeleteBackupConfig(deleteBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigDeleteRequest := backupClient.BackupConfigServiceDeleteBackupConfig(ctx, deleteBackupConfigRequest.BackupConfig.V1.Delete.Id)
	err = utilities.CopyStruct(backupConfigDeleteRequest, deleteBackupConfigRequest.BackupConfig.V1.Delete)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	_, res, err := backupClient.BackupConfigServiceDeleteBackupConfigExecute(backupConfigDeleteRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceDeleteBackupConfigExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil, err

}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PDSV2_API) ListBackupConfig(listBackupConfigRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	backupConfigResponse := []apiStructs.WorkFlowResponse{}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	backupConfigListRequest := backupClient.BackupConfigServiceListBackupConfigs(ctx)
	err = utilities.CopyStruct(backupConfigListRequest, listBackupConfigRequest.BackupConfig.V1.List)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	backupConfigModel, res, err := backupClient.BackupConfigServiceListBackupConfigsExecute(backupConfigListRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceListBackupConfigsExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(&backupConfigResponse, backupConfigModel)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while copying structs: %v\n", err)
	}

	return backupConfigResponse, err

}
