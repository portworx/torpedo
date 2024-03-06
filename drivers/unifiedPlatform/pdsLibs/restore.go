package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type WorkflowRestore struct {
	ProjectId    string
	DeploymentID string
	NamepsaceID  string
}

// CreateRestore creates restore for the backup
func (restore WorkflowRestore) CreateRestore() (*apiStructs.WorkFlowResponse, error) {

	createRestoreRequest := apiStructs.WorkFlowRequest{}

	createRestoreRequest.Restore.Create.SourceReferences.BackupId = "SomeBackupID"
	createRestoreRequest.Restore.Create.DestinationReferences.TargetClusterId = "SomeClusterID"

	restoreResponse, err := v2Components.PDS.CreateRestore(&createRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// ReCreateRestore recreates restore of the deployment
func (restore WorkflowRestore) ReCreateRestore() (*apiStructs.WorkFlowResponse, error) {

	recreateRestore := apiStructs.WorkFlowRequest{}

	recreateRestore.Restore.ReCreate.Id = "SomeID"

	restoreResponse, err := v2Components.PDS.ReCreateRestore(&recreateRestore)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// DeleteRestore deletes restore of the deployment
func (restore WorkflowRestore) DeleteRestore() (*apiStructs.WorkFlowResponse, error) {

	deleteRestoreRequest := apiStructs.WorkFlowRequest{}

	deleteRestoreRequest.Restore.Delete.Id = "SomeRestoreID"

	restoreResponse, err := v2Components.PDS.DeleteRestore(&deleteRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// GetBackupConfig fetches backup config for the deployment
func (restore WorkflowRestore) GetRestore() (*apiStructs.WorkFlowResponse, error) {

	getRestoreRequest := apiStructs.WorkFlowRequest{}

	getRestoreRequest.Restore.Get.Id = "SomeRestoreID"

	restoreResponse, err := v2Components.PDS.GetRestore(&getRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// ListBackupConfig lists backup config for the deployment
func (restore WorkflowRestore) ListRestore() ([]apiStructs.WorkFlowResponse, error) {

	listRestoreRequest := apiStructs.WorkFlowRequest{}

	listRestoreRequest.Restore.List.Sort = &apiStructs.Sort{
		SortBy:    apiStructs.SortBy_Field(int32(90)),
		SortOrder: apiStructs.SortOrder_Value(int32(15)),
	}

	restoreResponse, err := v2Components.PDS.ListRestore(&listRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}
