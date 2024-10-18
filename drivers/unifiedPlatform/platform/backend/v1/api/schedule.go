package api

import (
	"fmt"
	api "github.com/pure-px/platform-api-go-client/platform/v1/backuppolicy"
	. "github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/pure-px/torpedo/drivers/utilities"
	"github.com/pure-px/torpedo/pkg/log"
	status "net/http"
)

// GetBackupPolicy gets the backup policy model by its ID.
func (backupPolicy *PLATFORM_API_V1) GetBackupPolicy(getRequest *BackupPolicyRequest) (*BackupPolicyResponse, error) {
	ctx, backupPolicyClient, err := backupPolicy.getbackupPolicyClient()
	backupPolicyResponse := BackupPolicyResponse{
		Get: V1BackupPolicy{},
	}

	requestBody := backupPolicyClient.BackupPolicyServiceGetBackupPolicy(ctx, getRequest.Get.Id)
	backupPolicyModel, res, err := backupPolicyClient.BackupPolicyServiceGetBackupPolicyExecute(requestBody)

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupPolicyServiceGetBackupPolicy`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(backupPolicyModel, &backupPolicyResponse.Get)

	return &backupPolicyResponse, nil
}

// ListBackupPolicy lists all backup policies.
func (backupPolicy *PLATFORM_API_V1) ListBackupPolicy(listRequest *BackupPolicyRequest) (*BackupPolicyResponse, error) {
	ctx, backupPolicyClient, err := backupPolicy.getbackupPolicyClient()
	backupPolicyResponse := BackupPolicyResponse{
		List: ListBackupPolicyResponse{},
	}

	requestBody := backupPolicyClient.BackupPolicyServiceListBackupPolicies(ctx)

	if listRequest.List.SortSortBy != "" {
		requestBody = requestBody.SortSortBy(listRequest.List.SortSortBy)
	}
	if listRequest.List.SortSortOrder != "" {
		requestBody = requestBody.SortSortOrder(listRequest.List.SortSortOrder)
	}
	if listRequest.List.PaginationPageNumber != "" {
		requestBody = requestBody.PaginationPageNumber(listRequest.List.PaginationPageNumber)
	}
	if listRequest.List.PaginationPageSize != "" {
		requestBody = requestBody.PaginationPageSize(listRequest.List.PaginationPageSize)
	}

	backupPolicyModel, res, err := backupPolicyClient.BackupPolicyServiceListBackupPoliciesExecute(requestBody)

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupPolicyServiceListBackupPoliciesExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(backupPolicyModel, &backupPolicyResponse.List)

	return &backupPolicyResponse, nil
}

// CreateBackupPolicy creates a backup policy.
func (backupPolicy *PLATFORM_API_V1) CreateBackupPolicy(createRequest *BackupPolicyRequest) (*BackupPolicyResponse, error) {
	ctx, backupPolicyClient, err := backupPolicy.getbackupPolicyClient()
	backupPolicyResponse := BackupPolicyResponse{
		List: ListBackupPolicyResponse{},
	}

	requestBody := backupPolicyClient.BackupPolicyServiceCreateBackupPolicy(ctx, createRequest.Create.TenantId)
	policyBody := api.V1BackupPolicy{
		Meta: &api.V1Meta{
			Name: createRequest.Create.Meta.Name,
		},
		Config: &api.V1Config{
			Schedule: []api.V1Schedule{},
		},
	}

	for _, eachSchedule := range createRequest.Create.Config.Schedule {
		schedule := api.V1Schedule{
			// TODO: Add other policies if required
			IntervalPolicy: &api.V1IntervalPolicy{
				Minutes: eachSchedule.IntervalPolicy.Minutes,
			},
			CronExpression:   eachSchedule.CronExpression,
			IncrementalCount: eachSchedule.IncrementalCount,
			Retain:           eachSchedule.Retain,
		}

		policyBody.Config.Schedule = append(policyBody.Config.Schedule, schedule)
	}

	requestBody = requestBody.V1BackupPolicy(policyBody)
	log.Infof("Request body: %+v", requestBody)

	backupPolicyModel, res, err := backupPolicyClient.BackupPolicyServiceCreateBackupPolicyExecute(requestBody)

	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupPolicyServiceCreateBackupPolicyExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(backupPolicyModel, &backupPolicyResponse.Create)

	return &backupPolicyResponse, nil
}

// DeleteBackupPolicy deletes a backup policy.
func (backupPolicy *PLATFORM_API_V1) DeleteBackupPolicy(deleteRequest *BackupPolicyRequest) error {
	ctx, backupPolicyClient, err := backupPolicy.getbackupPolicyClient()

	requestBody := backupPolicyClient.BackupPolicyServiceGetBackupPolicy(ctx, deleteRequest.Delete.Id)
	_, res, err := backupPolicyClient.BackupPolicyServiceGetBackupPolicyExecute(requestBody)

	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `BackupPolicyServiceGetBackupPolicyExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	return nil
}
