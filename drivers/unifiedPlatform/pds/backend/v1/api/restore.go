package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// CreateRestore will create restore for a given backup
func (restore *PDSV2_API) CreateRestore(createRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("CreateRestore is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// ReCreateRestore will recreate restore for a given deployment
func (restore *PDSV2_API) ReCreateRestore(recretaeRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("ReCreateRestore is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// GetRestore will fetch restore for a given deployment
func (restore *PDSV2_API) GetRestore(getRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("GetRestore is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// DeleteRestore will delete restore for a given deployment
func (restore *PDSV2_API) DeleteRestore(deleteRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("DeleteRestore is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}

// ListRestore will list restores for a given deployment
func (restore *PDSV2_API) ListRestore(listRestoresRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Warnf("DeleteRestore is not implemented for API")
	return []apiStructs.WorkFlowResponse{}, nil
}
