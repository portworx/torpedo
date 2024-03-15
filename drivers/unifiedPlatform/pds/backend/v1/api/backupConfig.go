package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	backupConfigV1 "github.com/pure-px/platform-api-go-client/pds/v1/backupconfig"
)

var (
	BackupRequestBody backupConfigV1.V1BackupConfig
)

// CreateBackupConfig will create backup config for a given deployment
func (backupConf *PDS_API_V1) CreateBackupConfig(createBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PDS_API_V1) UpdateBackupConfig(updateBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("UpdateBackupConfig is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PDS_API_V1) GetBackupConfig(getBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("GetBackupConfig is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PDS_API_V1) DeleteBackupConfig(deleteBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("DeleteBackupConfig is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PDS_API_V1) ListBackupConfig(listBackupConfigRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Warnf("ListBackupConfig is not implemented for API")
	return []apiStructs.WorkFlowResponse{}, nil

}
