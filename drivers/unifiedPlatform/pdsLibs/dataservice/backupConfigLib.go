package dataservice

import (
	"context"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type BackupConfig struct {
	ProjectId        string
	DeploymentID     string
	BackupConfigType *pdsv2.ConfigBackupType
	BackupLevel      *pdsv2.ConfigBackupLevel
	ReclaimPolicy    *pdsv2.ConfigReclaimPolicyType
}

func CreateBackupConfig(backupConfig BackupConfig, ctx context.Context) (*apiStructs.WorkFlowResponse, error) {

	createBackupRequest := apiStructs.WorkFlowRequest{}

	createBackupRequest.BackupConfig.Create.V1 = createBackupRequest.BackupConfig.Create.V1.ApiService.
		BackupConfigServiceCreateBackupConfig(ctx, backupConfig.ProjectId)

	createBackupRequest.BackupConfig.Create.V1.DeploymentId(backupConfig.DeploymentID)
	createBackupRequest.BackupConfig.Create.V1.V1BackupConfig(pdsv2.V1BackupConfig{
		Config: &pdsv2.V1Config{
			JobHistoryLimit: intToPointerInt(5),
			Schedule: &pdsv2.V1Schedule{
				Id: intToPointerString(1),
			},
			Suspend:       PointerBool(false),
			BackupType:    backupConfig.BackupConfigType,
			BackupLevel:   backupConfig.BackupLevel,
			ReclaimPolicy: backupConfig.ReclaimPolicy,
		},
	})

	backupResponse, err := v2Components.PDS.CreateBackupConfig(&createBackupRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}
