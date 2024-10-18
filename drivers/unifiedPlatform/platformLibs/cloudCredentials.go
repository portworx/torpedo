package platformLibs

import (
	"fmt"
	"github.com/pure-px/torpedo/pkg/log"
	"os"
	"strconv"
	"strings"

	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	utilities "github.com/pure-px/torpedo/drivers/utilities"
)

var (
	key      string
	value    string
	provider int32
)

func GetCloudCredentials(credId, backupType string, isConfigRequired bool) (*automationModels.CloudCredentialsResponse, error) {
	getReq := automationModels.CloudCredentialsRequest{}
	getReq.Get.CloudCredentialsId = credId
	getReq.Get.IsConfigRequired = isConfigRequired

	switch backupType {
	case "s3":
		getReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
	case "azure":
		getReq.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
	case "gcp":
		getReq.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
	case "s3-comp":
		getReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
	default:
		getReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
	}

	wfResponse, err := v2Components.Platform.GetCloudCredentials(&getReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudcredentials: %v\n", err)
	}

	return wfResponse, nil
}

func ListCloudCredential(tenantId, label, sortBy, sortOrder string) (*automationModels.CloudCredentialsResponse, error) {
	req := automationModels.CloudCredentialsRequest{
		List: automationModels.CloudCredentials{
			TenantID:      tenantId,
			Label:         label,
			SortSortBy:    sortBy,
			SortSortOrder: sortOrder,
		},
	}

	cloudCredentials, err := v2Components.Platform.ListCloudCredentials(&req)
	if err != nil {
		return nil, fmt.Errorf("Error while listing cloud credentials %v\n", err)
	}
	totalPages, err := strconv.Atoi(*cloudCredentials.List.Pagination.TotalPages)
	if err != nil {
		return cloudCredentials, fmt.Errorf("Unable to get total pages")
	}
	totalRecords, err := strconv.Atoi(*cloudCredentials.List.Pagination.TotalRecords)
	if err != nil {
		return cloudCredentials, fmt.Errorf("Unable to get total records")
	}

	req.List.PaginationPageNumber = "1"
	req.List.PaginationPageSize = *cloudCredentials.List.Pagination.TotalRecords

	cloudCredentials, err = v2Components.Platform.ListCloudCredentials(&req)
	if err != nil {
		return nil, fmt.Errorf("Error while listing cloud credentials %v\n", err)
	}
	log.Infof("Cloud Credentials have [%d] pages and [%d] records", totalPages, totalRecords)

	return cloudCredentials, nil
}

func DeleteCloudCredential(cloudCredentialsId string) error {
	req := automationModels.CloudCredentialsRequest{}
	req.Get.CloudCredentialsId = cloudCredentialsId
	err := v2Components.Platform.DeleteCloudCredential(&req)
	if err != nil {
		return fmt.Errorf("failed to delete cloudcredentials: %v\n", err)
	}
	return nil
}

func CreateCloudCredentials(tenantId, backupType string) (*automationModels.CloudCredentialsResponse, error) {
	createReq := automationModels.CloudCredentialsRequest{}
	credsName := strings.ToLower("pds-bkp-creds-" + utilities.RandString(5))
	createReq.Create.TenantID = tenantId
	createReq.Create.Meta.Name = &credsName

	switch backupType {
	case "s3":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.S3Credentials.AccessKey = os.Getenv(envAwsAccessKey)
		createReq.Create.Config.S3Credentials.SecretKey = os.Getenv(envAwsSecretKey)
	case "azure":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.Create.Config.AzureCredentials.AccountKey = os.Getenv(envAzurePrimaryAccountKey)
		createReq.Create.Config.AzureCredentials.AccountName = os.Getenv(envAzureStorageAccountName)
	case "gcp":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_GOOGLE
		createReq.Create.Config.GoogleCredentials.ProjectId = os.Getenv(envGcpProjectId)
		createReq.Create.Config.GoogleCredentials.Key = os.Getenv(envGcpJsonPath)
	case "s3-comp":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.S3Credentials.AccessKey = os.Getenv(envMinioAccessKey)
		createReq.Create.Config.S3Credentials.SecretKey = os.Getenv(envMinioSecretKey)
	default:
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.S3Credentials.AccessKey = os.Getenv(envMinioAccessKey)
		createReq.Create.Config.S3Credentials.SecretKey = os.Getenv(envMinioSecretKey)
	}

	wfResponse, err := v2Components.Platform.CreateCloudCredentials(&createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudcredentials: %v\n", err)
	}
	return wfResponse, nil
}
