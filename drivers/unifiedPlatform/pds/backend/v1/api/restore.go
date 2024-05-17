package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	restoreV1 "github.com/pure-px/platform-api-go-client/pds/v1/restore"
	status "net/http"
)

// CreateRestore will create restore for a given backup
func (restore *PDS_API_V1) CreateRestore(createRestoreRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {

	response := automationModels.PDSRestoreResponse{
		Create: automationModels.PDSRestore{},
	}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	sourceReference := restoreV1.NewV1SourceReferences(createRestoreRequest.Create.SourceReferences.BackupId)

	restoreRequest := restoreClient.RestoreServiceCreateRestore(ctx, createRestoreRequest.Create.NamespaceId)
	restoreRequest = restoreRequest.RestoreServiceCreateRestoreBody(
		restoreV1.RestoreServiceCreateRestoreBody{
			ProjectId: createRestoreRequest.Create.ProjectId,
			Restore: restoreV1.V1Restore{
				Meta: &restoreV1.V1Meta{
					Name: createRestoreRequest.Create.Restore.Meta.Name,
				},
				Config: &restoreV1.V1Config{
					SourceReferences: sourceReference,
				},
			},
		},
	)

	restoreModel, res, err := restoreClient.RestoreServiceCreateRestoreExecute(restoreRequest)

	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceCreateRestore`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(restoreModel, &response.Create)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ReCreateRestore will recreate restore for a given deployment
func (restore *PDS_API_V1) ReCreateRestore(recretaeRestoreRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {
	response := automationModels.PDSRestoreResponse{
		ReCreate: automationModels.PDSRestore{},
	}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	restoreRequest := restoreClient.RestoreServiceRecreateRestore(ctx, recretaeRestoreRequest.ReCreate.NamespaceId)
	restoreRequest = restoreRequest.RestoreServiceRecreateRestoreBody(
		restoreV1.RestoreServiceRecreateRestoreBody{
			Name:        recretaeRestoreRequest.ReCreate.Name,
			ProjectId:   recretaeRestoreRequest.ReCreate.ProjectId,
			NamespaceId: recretaeRestoreRequest.ReCreate.NamespaceId,
		},
	)

	restoreModel, res, err := restoreRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceRecreateRestoreBody`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(restoreModel, &response.ReCreate)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetRestore will fetch restore for a given deployment
func (restore *PDS_API_V1) GetRestore(getRestoreRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {
	response := automationModels.PDSRestoreResponse{
		Get: automationModels.PDSRestore{},
	}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	restoreRequest := restoreClient.RestoreServiceGetRestore(ctx, getRestoreRequest.Get.Id)

	restoreModel, res, err := restoreClient.RestoreServiceGetRestoreExecute(restoreRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceGetRestore`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(restoreModel, &response.Get)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

//// DeleteRestore will delete restore for a given deployment
//func (restore *PDS_API_V1) DeleteRestore(deleteRestoreRequest *automationModels.PDSRestoreRequest) error {
//	ctx, restoreClient, err := restore.getRestoreClient()
//	if err != nil {
//		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
//	}
//
//	deleteRequest := restoreClient.(ctx, deleteRestoreRequest.Delete.Id)
//
//	_, res, err := deleteRequest.Execute()
//	if err != nil || res.StatusCode != status.StatusOK {
//		return fmt.Errorf("Error when calling `RestoreServiceGetRestore`: %v\n.Full HTTP response: %v", err, res)
//	}
//
//	return nil
//}

// ListRestore will list restores for a given deployment
func (restore *PDS_API_V1) ListRestore(listRestoresRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {
	response := automationModels.PDSRestoreResponse{
		List: automationModels.PDSListRestoreResponse{},
	}

	ctx, restoreClient, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	restoreRequest := restoreClient.RestoreServiceListRestores(ctx)
	restoreRequest = restoreRequest.DeploymentId(listRestoresRequest.List.DeploymentId)
	restoreRequest = restoreRequest.BackupId(listRestoresRequest.List.BackupId)
	restoreRequest = restoreRequest.ProjectId(listRestoresRequest.List.ProjectId)
	restoreRequest = restoreRequest.TenantId(listRestoresRequest.List.TenantId)

	restoreModels, res, err := restoreRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `RestoreServiceListRestores`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(restoreModels, &response.List)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
