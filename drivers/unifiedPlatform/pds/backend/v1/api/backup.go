package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// DeleteBackup will delete backup for a given deployment
func (backup *PDS_API_V1) DeleteBackup(deleteBackupRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {

	return &automationModels.WorkFlowResponse{}, nil
}

// ListBackup will list backup for a given deployment
func (backup *PDS_API_V1) ListBackup(listBackupConfigRequest *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Warnf("ListBackup is not implemented for API")
	return []automationModels.WorkFlowResponse{}, nil
}

// GetBackup will fetch backup for a given deployment
func (backup *PDS_API_V1) GetBackup(getBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("GetBackup is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}
