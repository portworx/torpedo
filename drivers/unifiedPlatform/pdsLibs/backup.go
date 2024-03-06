package pdslibs

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type WorkflowBackup struct {
	ProjectId        string
	DeploymentID     string
	BackupConfigType *pdsv2.ConfigBackupType
	BackupLevel      *pdsv2.ConfigBackupLevel
	ReclaimPolicy    *pdsv2.ConfigReclaimPolicyType
}

// DeleteBackup deletes backup config of the deployment
func DeleteBackup(backup WorkflowBackup) (*apiStructs.WorkFlowResponse, error) {

	deleteBackupRequest := apiStructs.WorkFlowRequest{}

	deleteBackupRequest.Backup.Delete.Id = "SomeID"

	backupResponse, err := v2Components.PDS.DeleteBackup(&deleteBackupRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// GetBackup fetches backup config for the deployment
func GetBackup(backup WorkflowBackup) (*apiStructs.WorkFlowResponse, error) {

	getBackupRequest := apiStructs.WorkFlowRequest{}

	getBackupRequest.Backup.Get.Id = "SomeID"

	backupResponse, err := v2Components.PDS.GetBackup(&getBackupRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// ListBackup lists backup config for the deployment
func ListBackup(backup WorkflowBackup) ([]apiStructs.WorkFlowResponse, error) {

	listBackup := apiStructs.WorkFlowRequest{}

	listBackup.Backup.List.Sort = &apiStructs.Sort{
		SortBy:    apiStructs.SortBy_Field(int32(25)),
		SortOrder: apiStructs.SortOrder_Value(int32(32)),
	}

	backupResponse, err := v2Components.PDS.ListBackup(&listBackup)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
