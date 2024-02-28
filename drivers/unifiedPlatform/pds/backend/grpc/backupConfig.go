package grpc

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// CreateBackupConfig will create backup config for a given deployment
func (backupConf *PdsGrpc) CreateBackupConfig(createBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PdsGrpc) UpdateBackupConfig(updateBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("UpdateBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PdsGrpc) GetBackupConfig(getBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("GetBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PdsGrpc) DeleteBackupConfig(deleteBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("DeleteBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PdsGrpc) ListBackupConfig(listBackupConfigRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Warnf("ListBackupConfig is not implemented for GRPC")
	return []apiStructs.WorkFlowResponse{}, nil

}
