package platformLibs

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"os"
	"strconv"
	"strings"
)

func ListBackupLocation(tenantId, label, sortBy, sortOrder string) (*automationModels.BackupLocationResponse, error) {
	listReq := automationModels.BackupLocationRequest{
		List: automationModels.BackupLocation{
			TenantID:      tenantId,
			Label:         label,
			SortSortBy:    sortBy,
			SortSortOrder: sortOrder,
		},
	}

	bkpLocations, err := v2Components.Platform.ListBackupLocations(&listReq)
	if err != nil {
		return nil, fmt.Errorf("Error while listing backup locations %v\n", err)
	}

	totalPages, err := strconv.Atoi(*bkpLocations.List.Pagination.TotalPages)
	if err != nil {
		return bkpLocations, fmt.Errorf("Unable to get total pages")
	}
	totalRecords, err := strconv.Atoi(*bkpLocations.List.Pagination.TotalRecords)
	if err != nil {
		return bkpLocations, fmt.Errorf("Unable to get total records")
	}

	log.Infof("Backup Locations have [%d] pages and [%d] records", totalPages, totalRecords)

	listReq.List.PaginationPageNumber = "1"
	listReq.List.PaginationPageSize = *bkpLocations.List.Pagination.TotalRecords

	bkpLocations, err = v2Components.Platform.ListBackupLocations(&listReq)
	if err != nil {
		return nil, fmt.Errorf("Error while listing backup locations %v\n", err)
	}
	log.Infof("Total records found - [%d]", len(bkpLocations.List.BackupLocations))
	return bkpLocations, nil
}

func CreateBackupLocation(tenantId, cloudCredId, bucketName, bkpLocation string) (*automationModels.BackupLocationResponse, error) {
	createReq := automationModels.BackupLocationRequest{}
	bkpLocName := strings.ToLower("pds-bkp-loc-" + utilities.RandString(5))

	createReq.Create.TenantID = tenantId
	createReq.Create.Meta.Name = &bkpLocName
	switch bkpLocation {
	case "s3":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.CloudCredentialsId = cloudCredId
		createReq.Create.Config.BkpLocation.S3Storage.BucketName = bucketName
		createReq.Create.Config.BkpLocation.S3Storage.Region = os.Getenv(envAwsRegion)
		createReq.Create.Config.BkpLocation.S3Storage.Endpoint = os.Getenv(envMinioEndPoint)

	case "s3-comp":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.CloudCredentialsId = cloudCredId
		createReq.Create.Config.BkpLocation.S3Storage.BucketName = bucketName
		createReq.Create.Config.BkpLocation.S3Storage.Region = os.Getenv(envMinioRegion)
		createReq.Create.Config.BkpLocation.S3Storage.Endpoint = os.Getenv(envMinioEndPoint)

	case "azure":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_AZURE
		createReq.Create.Config.CloudCredentialsId = cloudCredId
		createReq.Create.Config.BkpLocation.AzureStorage.ContainerName = bucketName

	case "gcp":
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_GOOGLE
		createReq.Create.Config.CloudCredentialsId = cloudCredId
		createReq.Create.Config.BkpLocation.GoogleStorage.BucketName = bucketName

	default:
		createReq.Create.Config.Provider.CloudProvider = PROVIDER_S3
		createReq.Create.Config.CloudCredentialsId = cloudCredId
		createReq.Create.Config.BkpLocation.S3Storage.BucketName = bucketName
		createReq.Create.Config.BkpLocation.S3Storage.Region = os.Getenv(envMinioRegion)
		createReq.Create.Config.BkpLocation.S3Storage.Endpoint = os.Getenv(envMinioEndPoint)

	}
	resp, err := v2Components.Platform.CreateBackupLocation(&createReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to create backup location: %v\n", err)
	}
	return resp, nil
}

func DeleteBackupLocation(bkpLocaitonId string) error {
	req := automationModels.BackupLocationRequest{
		Delete: automationModels.DeleteBackupLocation{
			Id: bkpLocaitonId,
		},
	}

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

// CreateAzureBucket creates bucket in Azure
func CreateAzureBucket(bucketName string) error {
	azureClient := utilities.AzureStorageClient{
		AccountName: utilities.GetEnv(envAzureStorageAccountName, envMinioAccessKey),
		AccountKey:  utilities.GetEnv(envAzurePrimaryAccountKey, envMinioSecretKey),
	}

	err := azureClient.CreateAzureBucket(bucketName)
	if err != nil {
		return fmt.Errorf("Failed to create bucket: %v\n", err)
	}
	return nil
}

// CreateGCPBucket creates bucket in gcp
func CreateGCPBucket(bucketName string) error {
	gcpClient := utilities.GcpStorageClient{
		ProjectId: utilities.GetEnv(envGcpProjectId, envMinioAccessKey),
	}

	err := SetGcpJsonPath()
	if err != nil {
		return err
	}
	err = CreateGcpJsonFile("/tmp/json")
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("/tmp/json"))
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	if _, err := client.Bucket(bucketName).Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			if err := client.Bucket(bucketName).Create(ctx, gcpClient.ProjectId, nil); err != nil {
				return fmt.Errorf("failed to create bucket: %v", err)
			}
			log.Infof("Bucket created: gs://%s\n", bucketName)
		} else {
			apiErr, ok := err.(*googleapi.Error)
			if ok && apiErr.Code == 403 {
				return fmt.Errorf("access denied to bucket: %v", err)
			} else {
				return fmt.Errorf("failed to get bucket: %v", err)
			}
		}
	} else {
		log.Infof("Bucket already exists: gs://%s\n", bucketName)
	}

	return nil
}

func SetGcpJsonPath() error {
	cm, err := core.Instance().GetConfigMap("custom-pds-qa-gcp-json-path", "default")
	if err != nil {
		return err
	}
	if _, ok := cm.Data["gcp_json"]; ok {
		gcpJsonData := cm.Data["gcp_json"]
		os.Setenv("GCP_JSON_PATH", gcpJsonData)
		return nil
	}
	return fmt.Errorf("key: custom-pds-qa-gcp-json-path doesn't exists in the gcp configmap")
}

func CreateGcpJsonFile(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error while creating the file -> %v", err)
	}
	defer f.Close()
	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("error truncating file. Err: %v", err)
	}
	_, err = f.WriteString(os.Getenv("GCP_JSON_PATH"))
	if err != nil {
		return fmt.Errorf("error while writing the data to file -> %v", err)
	}
	return nil
}
