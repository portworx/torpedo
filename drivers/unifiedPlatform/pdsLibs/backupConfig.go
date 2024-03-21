package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdsv2 "github.com/pure-px/platform-api-go-client/pds/v1/backupconfig"
)

type WorkflowBackupInput struct {
	ProjectId        string
	DeploymentID     string
	BackupConfigType *pdsv2.ConfigBackupType
	BackupLevel      *pdsv2.ConfigBackupLevel
	ReclaimPolicy    *pdsv2.ConfigReclaimPolicyType
}

// CreateBackupConfig created backup config for the deployment
func CreateBackupConfig(backupConfig WorkflowBackupInput) (*automationModels.WorkFlowResponse, error) {

	createBackupConfigRequest := automationModels.WorkFlowRequest{}

	createBackupConfigRequest.BackupConfig.Create.BackupConfig = &automationModels.V1BackupConfig{
		Meta: &automationModels.Meta{
			Uid: intToPointerString(10),
		},
		Config: &automationModels.Config{
			UserEmail: intToPointerString(15),
		},
		Status: &automationModels.Backupconfigv1Status{
			CustomResourceName: intToPointerString(70),
		},
	}
	createBackupConfigRequest.BackupConfig.Create.DeploymentId = backupConfig.DeploymentID
	createBackupConfigRequest.BackupConfig.Create.ProjectId = backupConfig.ProjectId

	backupResponse, err := v2Components.PDS.CreateBackupConfig(&createBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// UpdateBackupConfig updates backup config of the deployment
func UpdateBackupConfig(backupConfig WorkflowBackupInput) (*automationModels.WorkFlowResponse, error) {

	updateBackupConfigRequest := automationModels.WorkFlowRequest{}

	updateBackupConfigRequest.BackupConfig.Update.BackupConfig = &automationModels.V1BackupConfig{
		Meta: &automationModels.Meta{
			Uid: intToPointerString(10),
		},
		Config: &automationModels.Config{
			UserEmail: intToPointerString(15),
		},
		Status: &automationModels.Backupconfigv1Status{
			CustomResourceName: intToPointerString(70),
		},
	}

	backupResponse, err := v2Components.PDS.UpdateBackupConfig(&updateBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// DeleteBackupConfig deletes backup config of the deployment
func DeleteBackupConfig(backupConfig WorkflowBackupInput) (*automationModels.WorkFlowResponse, error) {

	deleteBackupConfigRequest := automationModels.WorkFlowRequest{}

	deleteBackupConfigRequest.BackupConfig.Delete.Id = "SomeID"

	backupResponse, err := v2Components.PDS.DeleteBackupConfig(&deleteBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// GetBackupConfig fetches backup config for the deployment
func GetBackupConfig(backupConfig WorkflowBackupInput) (*automationModels.WorkFlowResponse, error) {

	getBackupConfigRequest := automationModels.WorkFlowRequest{}

	getBackupConfigRequest.BackupConfig.Get.Id = "SomeID"

	backupResponse, err := v2Components.PDS.GetBackupConfig(&getBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// ListBackupConfig lists backup config for the deployment
func ListBackupConfig(backupConfig WorkflowBackupInput) ([]automationModels.WorkFlowResponse, error) {

	listBackupConfig := automationModels.WorkFlowRequest{}

	listBackupConfig.BackupConfig.List.Sort = &automationModels.Sort{
		SortBy:    automationModels.SortBy_Field(int32(90)),
		SortOrder: automationModels.SortOrder_Value(int32(15)),
	}

	backupResponse, err := v2Components.PDS.ListBackupConfig(&listBackupConfig)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
