package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	backupConfigV1 "github.com/pure-px/platform-api-go-client/pds/v1/backupconfig"
	status "net/http"
)

var (
	BackupRequestBody backupConfigV1.V1BackupConfig
)

// CreateBackupConfig will create backup config for a given deployment
func (backupConf *PDS_API_V1) CreateBackupConfig(createBackupConfigRequest *automationModels.PDSBackupConfigRequest) (*automationModels.PDSBackupConfigResponse, error) {
	response := automationModels.PDSBackupConfigResponse{
		Create: automationModels.V1BackupConfig{},
	}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	backupCreateRequest := backupClient.BackupConfigServiceCreateBackupConfig(ctx, createBackupConfigRequest.Create.ProjectId)
	backupCreateRequest = backupCreateRequest.BackupConfigServiceCreateBackupConfigBody(backupConfigV1.BackupConfigServiceCreateBackupConfigBody{
		DataServiceDeploymentId: createBackupConfigRequest.Create.DeploymentId,
		BackupConfig: backupConfigV1.V1BackupConfig{
			Meta: &backupConfigV1.V1Meta{
				Name: createBackupConfigRequest.Create.BackupConfig.Meta.Name,
			},
			Config: &backupConfigV1.V1Config{
				References: &backupConfigV1.V1References{
					BackupLocationId: *createBackupConfigRequest.Create.BackupConfig.Config.References.BackupLocationId,
				},
				BackupType:    (*backupConfigV1.ConfigBackupType)(createBackupConfigRequest.Create.BackupConfig.Config.BackupType),
				Suspend:       createBackupConfigRequest.Create.BackupConfig.Config.Suspend,
				BackupLevel:   (*backupConfigV1.ConfigBackupLevel)(createBackupConfigRequest.Create.BackupConfig.Config.BackupLevel),
				ReclaimPolicy: (*backupConfigV1.ConfigReclaimPolicyType)(createBackupConfigRequest.Create.BackupConfig.Config.ReclaimPolicy),
			},
		},
	})

	backupModel, res, err := backupCreateRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceCreateBackupConfigBody`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(backupModel, &response.Create)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PDS_API_V1) UpdateBackupConfig(updateBackupConfigRequest *automationModels.PDSBackupConfigRequest) (*automationModels.PDSBackupConfigResponse, error) {
	response := automationModels.PDSBackupConfigResponse{
		Update: automationModels.V1BackupConfig{},
	}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	backupUpdateRequest := backupClient.BackupConfigServiceUpdateBackupConfig(ctx, updateBackupConfigRequest.Update.ID)
	backupUpdateRequest = backupUpdateRequest.BackupConfigServiceUpdateBackupConfigBody(backupConfigV1.BackupConfigServiceUpdateBackupConfigBody{
		Labels:      updateBackupConfigRequest.Update.Labels,
		Annotations: updateBackupConfigRequest.Update.Annotations,
	})

	backupModel, res, err := backupUpdateRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceUpdateBackupConfigBody`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(backupModel, response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PDS_API_V1) GetBackupConfig(getBackupConfigRequest *automationModels.PDSBackupConfigRequest) (*automationModels.PDSBackupConfigResponse, error) {
	response := automationModels.PDSBackupConfigResponse{
		Get: automationModels.V1BackupConfig{},
	}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	backupGetRequest := backupClient.BackupConfigServiceGetBackupConfig(ctx, getBackupConfigRequest.Get.Id)

	backupModel, res, err := backupGetRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceGetBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(backupModel, &response.Get)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PDS_API_V1) DeleteBackupConfig(deleteBackupConfigRequest *automationModels.PDSBackupConfigRequest) error {

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	backupDeleteRequest := backupClient.BackupConfigServiceDeleteBackupConfig(ctx, deleteBackupConfigRequest.Delete.Id)

	_, res, err := backupClient.BackupConfigServiceDeleteBackupConfigExecute(backupDeleteRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `BackupConfigServiceDeleteBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil
}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PDS_API_V1) ListBackupConfig(listBackupConfigRequest *automationModels.PDSBackupConfigRequest) (*automationModels.PDSBackupConfigResponse, error) {
	response := automationModels.PDSBackupConfigResponse{
		List: automationModels.ListPDSBackupResponse{},
	}

	ctx, backupClient, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	backupListRequest := backupClient.BackupConfigServiceListBackupConfigs(ctx)
	backupListRequest = backupListRequest.TenantId(listBackupConfigRequest.List.TenantId)

	if listBackupConfigRequest.List.SortSortBy != "" {
		backupListRequest = backupListRequest.SortSortBy(listBackupConfigRequest.List.SortSortBy)
	}
	if listBackupConfigRequest.List.SortSortOrder != "" {
		backupListRequest = backupListRequest.SortSortOrder(listBackupConfigRequest.List.SortSortOrder)
	}
	if listBackupConfigRequest.List.PaginationPageNumber != "" {
		backupListRequest = backupListRequest.PaginationPageNumber(listBackupConfigRequest.List.PaginationPageNumber)
	}
	if listBackupConfigRequest.List.PaginationPageSize != "" {
		backupListRequest = backupListRequest.PaginationPageSize(listBackupConfigRequest.List.PaginationPageSize)
	}

	backupModels, res, err := backupListRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupConfigServiceGetBackupConfig`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(backupModels, &response.List)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
