package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// DeleteBackup will delete backup for a given deployment
func (backup *PDSV2_API) DeleteBackup(deleteBackupRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("DeleteBackup is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// ListBackup will list backup for a given deployment
func (backup *PDSV2_API) ListBackup(listBackupConfigRequest *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Warnf("ListBackup is not implemented for API")
	return []automationModels.WorkFlowResponse{}, nil
}

// GetBackup will fetch backup for a given deployment
func (backup *PDSV2_API) GetBackup(getBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("GetBackup is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}
