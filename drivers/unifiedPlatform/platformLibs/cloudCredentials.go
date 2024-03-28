package platformLibs

import (
	"fmt"
	"os"
	"strings"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	utilities "github.com/portworx/torpedo/drivers/utilities"
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
		createReq.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv(envAwsAccessKey)
		createReq.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv(envAwsSecretKey)
	case "azure":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.Create.Config.Credentials.AzureCredentials.AccountKey = os.Getenv(envAzurePrimaryAccountKey)
		createReq.Create.Config.Credentials.AzureCredentials.AccountName = os.Getenv(envAzureStorageAccountName)
	case "gcp":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.Create.Config.Credentials.GcpCredentials.ProjectId = os.Getenv(envGcpProjectId)
		createReq.Create.Config.Credentials.GcpCredentials.Key = os.Getenv(envGcpJsonPath)
	case "s3-comp":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv(envMinioAccessKey)
		createReq.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv(envMinioSecretKey)
	default:
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv(envMinioAccessKey)
		createReq.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv(envMinioSecretKey)
	}

	wfResponse, err := v2Components.Platform.CreateCloudCredentials(&createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudcredentials: %v\n", err)
	}
	return wfResponse, nil
}
