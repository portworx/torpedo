package grpc

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// GetBackupConfig will fetch backup for a given backup config
func (backup *PdsGrpc) GetBackup(getBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("GetBackup is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil
}

// DeleteBackupConfig will delete backup for a given backup config
func (backup *PdsGrpc) DeleteBackup(deleteBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("DeleteBackup is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil
}

// ListBackupConfig will list backup for a given backup config
func (backup *PdsGrpc) ListBackup(listBackupRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Warnf("ListBackup is not implemented for GRPC")
	return []apiStructs.WorkFlowResponse{}, nil

}
