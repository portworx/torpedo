package pdslibs

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type WorkflowBackupInput struct {
	ProjectId        string
	DeploymentID     string
	BackupConfigType *pdsv2.ConfigBackupType
	BackupLevel      *pdsv2.ConfigBackupLevel
	ReclaimPolicy    *pdsv2.ConfigReclaimPolicyType
}

// CreateBackupConfig created backup config for the deployment
func CreateBackupConfig(backupConfig WorkflowBackupInput) (*apiStructs.WorkFlowResponse, error) {

	createBackupConfigRequest := apiStructs.WorkFlowRequest{}

	createBackupConfigRequest.BackupConfig.Create.BackupConfig = &apiStructs.V1BackupConfig{
		Meta: &apiStructs.Meta{
			Uid: intToPointerString(10),
		},
		Config: &apiStructs.Config{
			UserEmail: intToPointerString(15),
		},
		Status: &apiStructs.Backupconfigv1Status{
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
func UpdateBackupConfig(backupConfig WorkflowBackupInput) (*apiStructs.WorkFlowResponse, error) {

	updateBackupConfigRequest := apiStructs.WorkFlowRequest{}

	updateBackupConfigRequest.BackupConfig.Update.BackupConfig = &apiStructs.V1BackupConfig{
		Meta: &apiStructs.Meta{
			Uid: intToPointerString(10),
		},
		Config: &apiStructs.Config{
			UserEmail: intToPointerString(15),
		},
		Status: &apiStructs.Backupconfigv1Status{
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
func DeleteBackupConfig(backupConfig WorkflowBackupInput) (*apiStructs.WorkFlowResponse, error) {

	deleteBackupConfigRequest := apiStructs.WorkFlowRequest{}

	deleteBackupConfigRequest.BackupConfig.Delete.Id = "SomeID"

	backupResponse, err := v2Components.PDS.DeleteBackupConfig(&deleteBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// GetBackupConfig fetches backup config for the deployment
func GetBackupConfig(backupConfig WorkflowBackupInput) (*apiStructs.WorkFlowResponse, error) {

	getBackupConfigRequest := apiStructs.WorkFlowRequest{}

	getBackupConfigRequest.BackupConfig.Get.Id = "SomeID"

	backupResponse, err := v2Components.PDS.GetBackupConfig(&getBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// ListBackupConfig lists backup config for the deployment
func ListBackupConfig(backupConfig WorkflowBackupInput) ([]apiStructs.WorkFlowResponse, error) {

	listBackupConfig := apiStructs.WorkFlowRequest{}

	listBackupConfig.BackupConfig.List.Sort = &apiStructs.Sort{
		SortBy:    apiStructs.SortBy_Field(int32(90)),
		SortOrder: apiStructs.SortOrder_Value(int32(15)),
	}

	backupResponse, err := v2Components.PDS.ListBackupConfig(&listBackupConfig)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
