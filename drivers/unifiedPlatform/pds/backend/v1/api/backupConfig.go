package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	backupConfigV1 "github.com/pure-px/platform-api-go-client/pds/v1/backupconfig"
)

var (
	BackupRequestBody backupConfigV1.V1BackupConfig
)

// CreateBackupConfig will create backup config for a given deployment
func (backupConf *PDS_API_V1) CreateBackupConfig(createBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PDS_API_V1) UpdateBackupConfig(updateBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("UpdateBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PDS_API_V1) GetBackupConfig(getBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("GetBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PDS_API_V1) DeleteBackupConfig(deleteBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("DeleteBackupConfig is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PDS_API_V1) ListBackupConfig(listBackupConfigRequest *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Warnf("ListBackupConfig is not implemented for API")
	return []automationModels.WorkFlowResponse{}, nil

}
