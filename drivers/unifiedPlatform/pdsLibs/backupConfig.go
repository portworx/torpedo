package pdslibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

var (
	defaultBackupSuspend       = false
	defaultBackupType          = "ADHOC"
	defaultBackupLevel         = "SNAPSHOT"
	defaultBackupRetainPolicy  = "RETAIN"
	defaultBackJobHistoryLimit = int32(10)
)

// CreateBackupConfig created backup config for the deployment
func CreateBackupConfig(name string, deploymentId string, projectId string, backupLocationId string) (*automationModels.PDSBackupConfigResponse, error) {

	createBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	createBackupConfigRequest.Create.BackupConfig = &automationModels.V1BackupConfig{
		Meta: &automationModels.Meta{
			Name: &name,
		},
		Config: &automationModels.Config{
			References: &automationModels.References{
				BackupLocationId: &backupLocationId,
			},
			BackupType:      &defaultBackupType,
			Suspend:         &defaultBackupSuspend,
			BackupLevel:     &defaultBackupLevel,
			ReclaimPolicy:   &defaultBackupRetainPolicy,
			JobHistoryLimit: &defaultBackJobHistoryLimit,
		},
	}
	createBackupConfigRequest.Create.DeploymentId = deploymentId
	createBackupConfigRequest.Create.ProjectId = projectId
	backupResponse, err := v2Components.PDS.CreateBackupConfig(&createBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// UpdateBackupConfig updates backup config of the deployment
func UpdateBackupConfig(id string, labels map[string]string, annotations map[string]string) (*automationModels.PDSBackupConfigResponse, error) {

	updateBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	updateBackupConfigRequest.Update.ID = id
	updateBackupConfigRequest.Update.Labels = &labels
	updateBackupConfigRequest.Update.Annotations = &annotations

	backupResponse, err := v2Components.PDS.UpdateBackupConfig(&updateBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// DeleteBackupConfig deletes backup config of the deployment
func DeleteBackupConfig(id string) error {

	deleteBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	deleteBackupConfigRequest.Delete.Id = id

	err := v2Components.PDS.DeleteBackupConfig(&deleteBackupConfigRequest)
	if err != nil {
		return err
	}
	return nil
}

// GetBackupConfig fetches backup config for the deployment
func GetBackupConfig(id string) (*automationModels.PDSBackupConfigResponse, error) {

	getBackupConfigRequest := automationModels.PDSBackupConfigRequest{}

	getBackupConfigRequest.Get.Id = id

	backupResponse, err := v2Components.PDS.GetBackupConfig(&getBackupConfigRequest)
	if err != nil {
		return nil, err
	}
	return backupResponse, err
}

// ListBackupConfig lists backup config for the deployment
func ListBackupConfig(tenantId string) (*automationModels.PDSBackupConfigResponse, error) {

	listBackupConfig := automationModels.PDSBackupConfigRequest{
		List: automationModels.ListPDSBackupConfig{
			TenantId: tenantId,
		},
	}

	backupResponse, err := v2Components.PDS.ListBackupConfig(&listBackupConfig)
	if err != nil {
		return nil, err
	}

	totalRecords := *backupResponse.List.Pagination.TotalRecords
	log.Infof("Total backup configs under [%s] = [%s]", tenantId, totalRecords)

	listBackupConfig = automationModels.PDSBackupConfigRequest{
		List: automationModels.ListPDSBackupConfig{
			TenantId:             tenantId,
			PaginationPageNumber: DEFAULT_PAGE_NUMBER,
			PaginationPageSize:   totalRecords,
			SortSortBy:           DEFAULT_SORT_BY,
			SortSortOrder:        DEFAULT_SORT_ORDER,
		},
	}

	backupResponse, err = v2Components.PDS.ListBackupConfig(&listBackupConfig)
	if err != nil {
		return nil, err
	}

	return backupResponse, err
}