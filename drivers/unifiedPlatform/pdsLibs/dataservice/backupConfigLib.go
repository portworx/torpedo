package dataservice

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type BackupConfig struct {
	ProjectId        string
	DeploymentID     *string
	BackupConfigType *pdsv2.ConfigBackupType
	BackupLevel      *pdsv2.ConfigBackupLevel
	ReclaimPolicy    *pdsv2.ConfigReclaimPolicyType
}

// CreateBackupConfig created backup config for the deployment
func CreateBackupConfig(backupConfig BackupConfig) (*apiStructs.WorkFlowResponse, error) {

	createBackupRequest := apiStructs.WorkFlowRequest{}

	createBackupRequest.BackupConfig.V1.Create.V1BackupConfig = &apiStructs.V1BackupConfig{}
	createBackupRequest.BackupConfig.V1.Create.DeploymentId = backupConfig.DeploymentID
	createBackupRequest.BackupConfig.V1.Create.ProjectId = backupConfig.ProjectId

	backupResponse, err := v2Components.PDS.CreateBackupConfig(&createBackupRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
