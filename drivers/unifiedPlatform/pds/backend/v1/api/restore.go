package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
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
func (restore *PDS_API_V1) GetRestore(getRestoreRequest *automationModels.WorkFlowRequest) (*automationModels.Restore, error) {
	restoreResponse := automationModels.Restore{}
	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}

	restrModel, res, err := restoreClient.RestoreServiceGetRestore(ctx, getRestoreRequest.Restore.Get.Id).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceGetRestore`: %v\n.Full HTTP response: %v", err, res)
	}

	err = copier.Copy(restoreResponse, restrModel)
	if err != nil {
		return nil, fmt.Errorf("Error occured while copying the restore response: %v\n", err)
	}

	return &restoreResponse, nil
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
