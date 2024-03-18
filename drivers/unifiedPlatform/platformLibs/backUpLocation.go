package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"os"
	"strings"
)

func CreateBackupLocation(tenantId, cloudCredId, bucketName, bkpLocation string) (*automationModels.BackupLocation, error) {
	createReq := automationModels.BackupLocation{}
	bkpLocName := strings.ToLower("pds-bkp-loc-" + utilities.RandString(5))
	createReq.TenantID = tenantId
	createReq.Meta.Name = &bkpLocName
	switch bkpLocation {
	case "s3":
		createReq.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Config.CloudCredentialsId = cloudCredId
		createReq.Config.BkpLocation.S3Storage.BucketName = bucketName
		createReq.Config.BkpLocation.S3Storage.Region = os.Getenv(envAwsRegion)
		createReq.Config.BkpLocation.S3Storage.Endpoint = os.Getenv(envMinioEndPoint)

	case "s3-comp":
		createReq.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Config.CloudCredentialsId = cloudCredId
		createReq.Config.BkpLocation.S3Storage.BucketName = bucketName
		createReq.Config.BkpLocation.S3Storage.Region = os.Getenv(envMinioRegion)
		createReq.Config.BkpLocation.S3Storage.Endpoint = os.Getenv(envMinioEndPoint)

	case "azure":
		createReq.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.Config.BkpLocation.AzureStorage.ContainerName = bucketName

	case "gcp":
		createReq.Config.Provider.CloudProvider = PROVIDER_GOOGLE
		createReq.Config.BkpLocation.GoogleStorage.BucketName = bucketName

	default:
		createReq.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Config.CloudCredentialsId = cloudCredId
		createReq.Config.BkpLocation.S3Storage.BucketName = bucketName
		createReq.Config.BkpLocation.S3Storage.Region = os.Getenv(envMinioRegion)
		createReq.Config.BkpLocation.S3Storage.Endpoint = os.Getenv(envMinioEndPoint)

	}
	resp, err := v2Components.Platform.CreateBackupLocation(&createReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to create backup location: %v\n", err)
	}
	return resp, nil
}
