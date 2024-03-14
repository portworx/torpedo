package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	utilities "github.com/portworx/torpedo/drivers/utilities"
	"os"
	"strings"
)

var (
	key      string
	value    string
	provider int32
)

func GetCloudCredentials(credId, backupType string, isConfigRequired bool) (*apiStructs.WorkFlowResponse, error) {
	getReq := apiStructs.WorkFlowRequest{}
	getReq.CloudCredentials.Get.CloudCredentialsId = credId
	getReq.CloudCredentials.Get.IsConfigRequired = isConfigRequired

	switch backupType {
	case "s3":
		getReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
	case "azure":
		getReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
	case "gcp":
		getReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
	case "s3-comp":
		getReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
	default:
		getReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
	}

	wfResponse, err := v2Components.Platform.GetCloudCredentials(&getReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to create cloudcredentials: %v\n", err)
	}

	return wfResponse, nil
}

func CreateCloudCredentials(tenantId, backupType string) (*apiStructs.WorkFlowResponse, error) {
	createReq := apiStructs.WorkFlowRequest{}
	credsName := strings.ToLower("pds-bkp-creds-" + utilities.RandString(5))
	createReq.CloudCredentials.Create.TenantID = tenantId
	createReq.CloudCredentials.Create.Meta.Name = &credsName

	switch backupType {
	case "s3":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv(envAwsAccessKey)
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv(envAwsSecretKey)
	case "azure":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.CloudCredentials.Create.Config.Credentials.AzureCredentials.AccountKey = os.Getenv(envAzurePrimaryAccountKey)
		createReq.CloudCredentials.Create.Config.Credentials.AzureCredentials.AccountName = os.Getenv(envAzureStorageAccountName)
	case "gcp":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.CloudCredentials.Create.Config.Credentials.GcpCredentials.ProjectId = os.Getenv(envGcpProjectId)
		createReq.CloudCredentials.Create.Config.Credentials.GcpCredentials.Key = os.Getenv(envGcpJsonPath)
	case "s3-comp":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv(envMinioAccessKey)
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv(envMinioSecretKey)
	default:
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv(envMinioAccessKey)
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv(envMinioSecretKey)
	}

	wfResponse, err := v2Components.Platform.CreateCloudCredentials(&createReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to create cloudcredentials: %v\n", err)
	}
	return wfResponse, nil
}
