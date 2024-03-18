package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"os"
	"strings"
)

func ListBackupLocation(tenantId string) ([]*automationModels.BackupLocationResponse, error) {
	listReq := automationModels.BackupLocationRequest{}
	listReq.List.TenantID = tenantId
	bkpLocations, err := v2Components.Platform.ListBackupLocations(&listReq)
	if err != nil {
		return nil, fmt.Errorf("Error while listing backup locations %v\n", err)
	}
	return bkpLocations, nil
}

func CreateBackupLocation(tenantId, cloudCredId, bucketName, bkpLocation string) (*automationModels.BackupLocationResponse, error) {
	req := automationModels.BackupLocationRequest{}
	bkpLocName := strings.ToLower("pds-bkp-loc-" + utilities.RandString(5))
	createReq := req.Create
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
	resp, err := v2Components.Platform.CreateBackupLocation(&req)
	if err != nil {
		return nil, fmt.Errorf("Failed to create backup location: %v\n", err)
	}
	return resp, nil
}

func DeleteBackupLocation(bkpLocaitonId string) error {
	req := automationModels.BackupLocationRequest{}
	req.List.Meta.Uid = &bkpLocaitonId
	err := v2Components.Platform.DeleteBackupLocation(&req)
	if err != nil {
		return fmt.Errorf("Failed to delete backup location: %v\n", err)
	}
	return nil
}

func CreateS3CompBucket(bucketName string) error {
	awsS3CompClient := utilities.AwsCompatibleStorageClient{
		Endpoint:  utilities.GetEnv(envMinioEndPoint, envMinioEndPoint),
		AccessKey: utilities.GetEnv(envMinioAccessKey, envMinioAccessKey),
		SecretKey: utilities.GetEnv(envMinioSecretKey, envMinioSecretKey),
		Region:    utilities.GetEnv(envMinioRegion, envMinioRegion),
	}
	err := awsS3CompClient.CreateS3CompBucket(bucketName)
	if err != nil {
		return fmt.Errorf("Failed to create bucket: %v\n", err)
	}
	return nil
}

func CreateS3Bucket(bucketName string) error {
	awsS3Client := utilities.AwsStorageClient{
		AccessKey: utilities.GetEnv(envAwsAccessKey, envMinioAccessKey),
		SecretKey: utilities.GetEnv(envAwsSecretKey, envMinioSecretKey),
		Region:    utilities.GetEnv(envAwsRegion, envMinioRegion),
	}
	err := awsS3Client.CreateS3Bucket(bucketName)
	if err != nil {
		return fmt.Errorf("Failed to create bucket: %v\n", err)
	}
	return nil
}
