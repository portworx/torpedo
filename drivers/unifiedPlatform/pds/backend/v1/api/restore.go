package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// CreateRestore will create restore for a given backup
func (restore *PDS_API_V1) CreateRestore(createRestoreRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("CreateRestore is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// ReCreateRestore will recreate restore for a given deployment
func (restore *PDS_API_V1) ReCreateRestore(recretaeRestoreRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("ReCreateRestore is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// GetRestore will fetch restore for a given deployment
func (restore *PDS_API_V1) GetRestore(getRestoreRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("GetRestore is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// DeleteRestore will delete restore for a given deployment
func (restore *PDS_API_V1) DeleteRestore(deleteRestoreRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("DeleteRestore is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}

// ListRestore will list restores for a given deployment
func (restore *PDS_API_V1) ListRestore(listRestoresRequest *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Warnf("DeleteRestore is not implemented for API")
	return []automationModels.WorkFlowResponse{}, nil
}
