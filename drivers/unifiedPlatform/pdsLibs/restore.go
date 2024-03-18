package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

type WorkflowRestore struct {
	ProjectId    string
	DeploymentID string
	NamepsaceID  string
}

// CreateRestore creates restore for the backup
func (restore WorkflowRestore) CreateRestore() (*automationModels.WorkFlowResponse, error) {

	createRestoreRequest := automationModels.WorkFlowRequest{}

	createRestoreRequest.Restore.Create.SourceReferences = &automationModels.SourceReferences{
		BackupId: "BackupID",
	}
	createRestoreRequest.Restore.Create.DestinationReferences = &automationModels.DestinationReferences{
		TargetClusterId: "TargetClusterID",
	}
	createRestoreRequest.Restore.Create.SourceReferences.BackupId = "SomeBackupID"
	createRestoreRequest.Restore.Create.DestinationReferences.TargetClusterId = "SomeClusterID"

	restoreResponse, err := v2Components.PDS.CreateRestore(&createRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// ReCreateRestore recreates restore of the deployment
func (restore WorkflowRestore) ReCreateRestore() (*automationModels.WorkFlowResponse, error) {

	recreateRestore := automationModels.WorkFlowRequest{}

	recreateRestore.Restore.ReCreate.Id = "SomeID"

	restoreResponse, err := v2Components.PDS.ReCreateRestore(&recreateRestore)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// DeleteRestore deletes restore of the deployment
func (restore WorkflowRestore) DeleteRestore() (*automationModels.WorkFlowResponse, error) {

	deleteRestoreRequest := automationModels.WorkFlowRequest{}

	deleteRestoreRequest.Restore.Delete.Id = "SomeRestoreID"

	restoreResponse, err := v2Components.PDS.DeleteRestore(&deleteRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// GetBackupConfig fetches backup config for the deployment
func (restore WorkflowRestore) GetRestore() (*automationModels.WorkFlowResponse, error) {

	getRestoreRequest := automationModels.WorkFlowRequest{}

	getRestoreRequest.Restore.Get.Id = "SomeRestoreID"

	restoreResponse, err := v2Components.PDS.GetRestore(&getRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// ListBackupConfig lists backup config for the deployment
func (restore WorkflowRestore) ListRestore() ([]automationModels.WorkFlowResponse, error) {

	listRestoreRequest := automationModels.WorkFlowRequest{}

	listRestoreRequest.Restore.List.Sort = &automationModels.Sort{
		SortBy:    automationModels.SortBy_Field(int32(90)),
		SortOrder: automationModels.SortOrder_Value(int32(15)),
	}

	restoreResponse, err := v2Components.PDS.ListRestore(&listRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}
