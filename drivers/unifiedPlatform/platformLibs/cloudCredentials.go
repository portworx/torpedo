package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	utilities "github.com/portworx/torpedo/drivers/utilities"
	"os"
	"strings"
)

const (
	PROVIDER_UNSPECIFIED  int32 = 0
	PROVIDER_AZURE        int32 = 1
	PROVIDER_GOOGLE       int32 = 2
	PROVIDER_S3           int32 = 3
	PROVIDER_UNSTRUCTURED int32 = 4
)

var (
	key      string
	value    string
	provider int32
)

func GetCloudCredentials(credId, backupType string, isConfigRequired bool) (*automationModels.WorkFlowResponse, error) {
	getReq := automationModels.WorkFlowRequest{}
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

func CreateCloudCredentials(tenantId, backupType string) (*automationModels.WorkFlowResponse, error) {
	createReq := automationModels.WorkFlowRequest{}
	credsName := strings.ToLower("pds-automation-" + utilities.RandString(5))
	createReq.TenantId = tenantId
	createReq.CloudCredentials.Create.Meta.Name = &credsName

	switch backupType {
	case "s3":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	case "azure":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.CloudCredentials.Create.Config.Credentials.AzureCredentials.AccountKey = os.Getenv("AZURE_ACCOUNT_KEY")
		createReq.CloudCredentials.Create.Config.Credentials.AzureCredentials.AccountName = os.Getenv("AZURE_ACCOUNT_NAME")
	case "gcp":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.CloudCredentials.Create.Config.Credentials.GcpCredentials.ProjectId = os.Getenv("GCP_PROJECT_ID")
		createReq.CloudCredentials.Create.Config.Credentials.GcpCredentials.Key = os.Getenv("GCP_JSON_PATH")
	case "s3-comp":
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
	default:
		createReq.CloudCredentials.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.AccessKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
		createReq.CloudCredentials.Create.Config.Credentials.S3Credentials.SecretKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
	}

	wfResponse, err := v2Components.Platform.CreateCloudCredentials(&createReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to create cloudcredentials: %v\n", err)
	}
	return wfResponse, nil
}
