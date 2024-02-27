package dataservice

import (
	"context"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type Restore struct {
	DeploymentId        *string
	BackupId            *string
	BackupLocationId    *string
	CloudsnapId         *string
	TargetClusterId     *string
	CustomResourceNames *string
}

// CreateRestore creates a restore from a backup
func CreateRestore(restoreRequest Restore, ctx context.Context) (*apiStructs.WorkFlowResponse, error) {

	createRestoreRequest := apiStructs.WorkFlowRequest{}

	createRestoreRequest.Restore.Create.V1.ApiService.RestoreServiceCreateRestore(ctx, namespaceId)

	createRestoreRequest.Restore.Create.V1.V1Restore(pdsv2.V1Restore{
		Config: &pdsv2.V1Config3{
			SourceReferences: &pdsv2.V1SourceReferences{
				DeploymentId:     restoreRequest.DeploymentId,
				BackupId:         restoreRequest.BackupId,
				BackupLocationId: restoreRequest.BackupLocationId,
				CloudsnapId:      restoreRequest.CloudsnapId,
			},
			DestinationReferences: &pdsv2.V1DestinationReferences{
				TargetClusterId: restoreRequest.TargetClusterId,
				DeploymentId:    restoreRequest.DeploymentId,
			},
			CustomResourceName: restoreRequest.CustomResourceNames,
		},
	})

	restoreResponse, err := v2Components.PDS.CreateRestore(&createRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}
