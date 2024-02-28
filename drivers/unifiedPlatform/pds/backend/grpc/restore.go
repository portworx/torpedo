package grpc

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// CreateRestore will create restore from a given backup
func (restore *PdsGrpc) CreateRestore(createRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// ReCreateRestore will re-create restore from a given backup
func (restore *PdsGrpc) ReCreateRestore(recreateRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// GetRestore will fetch restore from existing restores
func (restore *PdsGrpc) GetRestore(getRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// DeleteRestore will delete restore from existing restores
func (restore *PdsGrpc) DeleteRestore(deleteRestoreRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {

	log.Warnf("CreateBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// ListRestore will list all restores
func (restore *PdsGrpc) ListRestore(listRestoreRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Warnf("CreateBackupConfig is not implemented for GRPC")
	return []apiStructs.WorkFlowResponse{}, nil

}
