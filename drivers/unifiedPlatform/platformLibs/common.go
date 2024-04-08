package platformLibs

import "github.com/portworx/torpedo/drivers/unifiedPlatform"

const (
	PROVIDER_UNSPECIFIED  int32 = 0
	PROVIDER_AZURE        int32 = 1
	PROVIDER_GOOGLE       int32 = 2
	PROVIDER_S3           int32 = 3
	PROVIDER_UNSTRUCTURED int32 = 4
)

const (
	// Backup environment variable
	envAwsAccessKey            = "AWS_ACCESS_KEY_ID"
	envAwsSecretKey            = "AWS_SECRET_ACCESS_KEY"
	envAwsRegion               = "AWS_REGION"
	envMinioAccessKey          = "AWS_MINIO_ACCESS_KEY_ID"
	envMinioSecretKey          = "AWS_MINIO_SECRET_ACCESS_KEY"
	envMinioRegion             = "AWS_MINIO_REGION"
	envMinioEndPoint           = "AWS_MINIO_ENDPOINT"
	envAzureStorageAccountName = "AZURE_ACCOUNT_NAME"
	envAzurePrimaryAccountKey  = "AZURE_ACCOUNT_KEY"
	envGcpProjectId            = "GCP_PROJECT_ID"
	envGcpJsonPath             = "GCP_JSON_PATH"
)

const (
	DEFAULT_PAGE_NUMBER = "1"
	DEFAULT_SORT_BY     = "CREATED_AT"
	DEFAULT_SORT_ORDER  = "DESC"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	err          error
)

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}
