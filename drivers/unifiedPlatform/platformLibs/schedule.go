package platformLibs

import "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"

func CreateBackupPolicy(name string, typeOfPolicy string, minutes string, retainCount string, tenantId string) (automationModels.V1BackupPolicy, error) {

	incremntalCount := "1" // Needs to be removed later as it is removed from backend
	// Create backup policy

	createRequest := automationModels.BackupPolicyRequest{
		Create: automationModels.CreateBackupPolicy{
			TenantId: tenantId,

			Meta: &automationModels.V1Meta{
				Name: &name,
			},
			Config: &automationModels.V1BackupPolicyConfig{
				Schedule: []automationModels.V1Schedule{},
			},
		},
	}

	schedule := automationModels.V1Schedule{}
	schedule.Retain = &retainCount
	schedule.IncrementalCount = &incremntalCount

	// TODO : Support needs to be added for other type when required

	switch typeOfPolicy {

	case automationModels.IntervalPolicy:
		schedule.IntervalPolicy = &automationModels.V1IntervalPolicy{
			Minutes: &minutes,
		}
	}

	createRequest.Create.Config.Schedule = append(createRequest.Create.Config.Schedule, schedule)

	// Call the API to create backup policy
	res, err := v2Components.Platform.CreateBackupPolicy(&createRequest)
	if err != nil {
		return automationModels.V1BackupPolicy{}, err
	}

	return res.Create, nil
}

func GetBackupPolicy(id string) (automationModels.V1BackupPolicy, error) {
	// Get backup policy

	getRequest := automationModels.BackupPolicyRequest{
		Get: automationModels.GetBackupPolicy{
			Id: id,
		},
	}

	// Call the API to get backup policy
	res, err := v2Components.Platform.GetBackupPolicy(&getRequest)
	if err != nil {
		return automationModels.V1BackupPolicy{}, err
	}

	return res.Get, nil
}

// DeleteBackupPolicy deletes a backup policy.
func DeleteBackupPolicy(id string) error {
	// Delete backup policy

	deleteRequest := automationModels.BackupPolicyRequest{
		Delete: automationModels.DeleteBackupPolicy{
			Id: id,
		},
	}

	// Call the API to delete backup policy
	err := v2Components.Platform.DeleteBackupPolicy(&deleteRequest)
	return err
}

// ListBackupPolicy lists all backup policies.
func ListBackupPolicy(pageNumber string, pageSize string, sortBy string, sortOrder string) (automationModels.ListBackupPolicyResponse, error) {
	// List backup policy

	listRequest := automationModels.BackupPolicyRequest{
		List: automationModels.ListBackupPolicy{
			PaginationPageNumber: pageNumber,
			PaginationPageSize:   pageSize,
			SortSortBy:           sortBy,
			SortSortOrder:        sortOrder,
		},
	}

	// Call the API to list backup policy
	res, err := v2Components.Platform.ListBackupPolicy(&listRequest)
	if err != nil {
		return automationModels.ListBackupPolicyResponse{}, err
	}

	return res.List, nil
}
