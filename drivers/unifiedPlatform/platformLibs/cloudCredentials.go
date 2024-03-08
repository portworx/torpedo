package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
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

func CreateCloudCredentials(tenantId, backupType string) (*apiStructs.WorkFlowResponse, error) {
	createReq := apiStructs.WorkFlowRequest{}
	credsName := strings.ToLower("pds-automation-" + utilities.RandString(5))
	createReq.TenantId = tenantId
	createReq.CloudCredentials.Meta.Name = &credsName

	switch backupType {
	case "s3":
		createReq.CloudCredentials.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Config.Credentials.S3Credentials.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
		createReq.CloudCredentials.Config.Credentials.S3Credentials.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	case "azure":
		createReq.CloudCredentials.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.CloudCredentials.Config.Credentials.AzureCredentials.AccountKey = os.Getenv("AZURE_ACCOUNT_KEY")
		createReq.CloudCredentials.Config.Credentials.AzureCredentials.AccountName = os.Getenv("AZURE_ACCOUNT_NAME")
	case "gcp":
		createReq.CloudCredentials.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.CloudCredentials.Config.Credentials.GcpCredentials.ProjectId = os.Getenv("GCP_PROJECT_ID")
		createReq.CloudCredentials.Config.Credentials.GcpCredentials.Key = os.Getenv("GCP_JSON_PATH")
	case "s3-comp":
		createReq.CloudCredentials.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Config.Credentials.S3Credentials.AccessKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
		createReq.CloudCredentials.Config.Credentials.S3Credentials.SecretKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
	default:
		createReq.CloudCredentials.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.CloudCredentials.Config.Credentials.S3Credentials.AccessKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
		createReq.CloudCredentials.Config.Credentials.S3Credentials.SecretKey = os.Getenv("AWS_MINIO_ACCESS_KEY_ID")
	}

	wfResponse, err := v2Components.Platform.CreateCloudCredentials(&createReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to create cloudcredentials: %v\n", err)
	}
	return wfResponse, nil
}
