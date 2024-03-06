package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// DeleteBackup will delete backup for a given deployment
func (backup *PDSV2_API) DeleteBackup(deleteBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("DeleteBackup is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// ListBackup will list backup for a given deployment
func (backup *PDSV2_API) ListBackup(listBackupConfigRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Warnf("ListBackup is not implemented for API")
	return []apiStructs.WorkFlowResponse{}, nil
}

// GetBackup will fetch backup for a given deployment
func (backup *PDSV2_API) GetBackup(getBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("GetBackup is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}
