package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

// CreateRestore creates restore for the backup
func CreateRestore(name string, backupId string, projectIdDest string) (*automationModels.PDSRestoreResponse, error) {

	createRestoreRequest := automationModels.PDSRestoreRequest{
		Create: automationModels.PDSCreateRestore{
			Restore: &automationModels.PDSRestore{
				Meta: &automationModels.Meta{
					Name: &name,
				},
			},
			ProjectId:   projectIdDest,
			NamespaceId: namespaceId,
			SourceReferences: &automationModels.SourceReferences{
				BackupId: backupId,
			},
			DestinationReferences: &automationModels.DestinationReferences{},
		},
	}

	restoreResponse, err := v2Components.PDS.CreateRestore(&createRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// ReCreateRestore recreates restore of the deployment
func ReCreateRestore(id string, targetClusterId string, name string, namespaceId string, projectId string) (*automationModels.PDSRestoreResponse, error) {

	recreateRestore := automationModels.PDSRestoreRequest{
		ReCreate: automationModels.PDSReCreateRestore{},
	}

	recreateRestore.ReCreate.Id = id
	recreateRestore.ReCreate.Name = name
	recreateRestore.ReCreate.ProjectId = projectId
	recreateRestore.ReCreate.TargetClusterId = targetClusterId
	recreateRestore.ReCreate.NamespaceId = namespaceId

	restoreResponse, err := v2Components.PDS.ReCreateRestore(&recreateRestore)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

//// DeleteRestore deletes restore of the deployment
//func DeleteRestore(id string) error {
//
//	deleteRestoreRequest := automationModels.PDSRestoreRequest{
//		Delete: automationModels.PDSDeleteRestore{},
//	}
//
//	deleteRestoreRequest.Delete.Id = id
//
//	err := v2Components.PDS.DeleteRestore(&deleteRestoreRequest)
//	if err != nil {
//		return err
//	}
//	return err
//}

// GetBackupConfig fetches backup config for the deployment
func GetRestore(id string) (*automationModels.PDSRestoreResponse, error) {

	getRestoreRequest := automationModels.PDSRestoreRequest{
		Get: automationModels.PDSGetRestore{
			Id: id,
		},
	}

	restoreResponse, err := v2Components.PDS.GetRestore(&getRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}

// ListBackupConfig lists backup config for the deployment
func ListRestore(accountId string, tenantId string, projectId string, deploymentId string, backupId string) (*automationModels.PDSRestoreResponse, error) {

	listRestoreRequest := automationModels.PDSRestoreRequest{
		List: automationModels.PDSListRestores{},
	}

	listRestoreRequest.List.AccountId = accountId
	listRestoreRequest.List.TenantId = tenantId
	listRestoreRequest.List.ProjectId = projectId
	listRestoreRequest.List.DeploymentId = deploymentId
	listRestoreRequest.List.BackupId = backupId

	restoreResponse, err := v2Components.PDS.ListRestore(&listRestoreRequest)
	if err != nil {
		return nil, err
	}
	return restoreResponse, err
}
