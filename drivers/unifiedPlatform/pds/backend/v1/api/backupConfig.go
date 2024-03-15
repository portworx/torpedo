package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
)

var (
	BackupRequestBody pdsv2.V1BackupConfig
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
func (backupConf *PDSV2_API) CreateBackupConfig(createBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PDSV2_API) UpdateBackupConfig(updateBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("UpdateBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PDSV2_API) GetBackupConfig(getBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("GetBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PDSV2_API) DeleteBackupConfig(deleteBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("DeleteBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PDSV2_API) ListBackupConfig(listBackupConfigRequest *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Warnf("ListBackupConfig is not implemented for API")
	return []automationModels.WorkFlowResponse{}, nil

}
