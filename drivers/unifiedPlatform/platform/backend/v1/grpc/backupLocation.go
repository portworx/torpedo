package grpc

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicbackuplocapi "github.com/pure-px/apis/public/portworx/platform/backuplocation/apiv1"
	publiccloudcredapi "github.com/pure-px/apis/public/portworx/platform/cloudcredential/apiv1"
	"google.golang.org/grpc"
)

type Provider_Type int32

// getBackupLocClient updates the header with bearer token and returns the new client
func (BackupLocGrpcV1 *PlatformGrpc) getBackupLocClient() (context.Context, publicbackuplocapi.BackupLocationServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var backupLocClient publicbackuplocapi.BackupLocationServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	ctx = WithAccountIDMetaCtx(ctx, BackupLocGrpcV1.AccountId)

	backupLocClient = publicbackuplocapi.NewBackupLocationServiceClient(BackupLocGrpcV1.ApiClientV1)

	return ctx, backupLocClient, token, nil
}

// ListBackupLocations return lis of backup locations
func (BackupLocGrpcV1 *PlatformGrpc) ListBackupLocations(request *BackupLocationRequest) (*BackupLocationResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bkpLocResponse := BackupLocationResponse{
		List: ListBackupLocation{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	listbkpLocationRequest := &publicbackuplocapi.ListBackupLocationsRequest{
		TenantId:   request.List.TenantID,
		Pagination: NewPaginationRequest(1, 50),
	}
	backupLocationModels, err := backupLocationClient.ListBackupLocations(ctx, listbkpLocationRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("Value of tenants - [%v]", backupLocationModels)

	for _, bkpLocation := range backupLocationModels.BackupLocations {
		log.Infof("printing cloud provider type [%s]", bkpLocation.Config.Provider.GetCloudProvider())
		resp := copyCloudLocationResponse(bkpLocation)
		bkpLocResponse.List.BackupLocations = append(bkpLocResponse.List.BackupLocations, resp.Create)
	}

	log.Infof("Value of backupLocation after copy - [%v]", bkpLocResponse)
	return &bkpLocResponse, nil
}

// GetBackupLocation get backup location model by its ID.
func (BackupLocGrpcV1 *PlatformGrpc) GetBackupLocation(getReq *WorkFlowRequest) (*BackupLocationResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	bckpLocResp := BackupLocationResponse{
		Get: BackupLocation{},
	}
	var getRequest *publicbackuplocapi.GetBackupLocationRequest
	err = copier.Copy(&getRequest, getReq)
	if err != nil {
		return nil, err
	}
	backupLocationModel, err := backupLocationClient.GetBackupLocation(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModel)
	//err = copier.Copy(&bckpLocResp, backupLocationModel)
	//if err != nil {
	//	return nil, err
	//}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// CreateBackupLocation return newly created backup location model.
func (BackupLocGrpcV1 *PlatformGrpc) CreateBackupLocation(createRequest *BackupLocationRequest) (*BackupLocationResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	createAppRequest := &publicbackuplocapi.CreateBackupLocationRequest{
		TenantId: createRequest.Create.TenantID,
		BackupLocation: &publicbackuplocapi.BackupLocation{
			Meta: &commonapiv1.Meta{
				Name: *createRequest.Create.Meta.Name,
			},
			Config: backupLocationConfig(createRequest),
		},
	}

	log.Infof("bucketName [%s]", createRequest.Create.Config.BkpLocation.S3Storage.BucketName)
	log.Infof("region [%s]", createRequest.Create.Config.BkpLocation.S3Storage.Region)
	log.Infof("endpoint [%s]", createRequest.Create.Config.BkpLocation.S3Storage.Endpoint)

	backupLocationModel, err := backupLocationClient.CreateBackupLocation(ctx, createAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `BackupLocationServiceCreateBackupLocation` to create backup target - %v", err)
	}

	log.Infof("Value of backupLocation - [%v]", backupLocationModel)

	bkpLocationResponse := copyCloudLocationResponse(backupLocationModel)

	log.Infof("Value of backupLocation after copy - [%v]", bkpLocationResponse)
	return bkpLocationResponse, nil
}

// UpdateBackupLocation return updated backup location model.
func (BackupLocGrpcV1 *PlatformGrpc) UpdateBackupLocation(updateRequest *WorkFlowRequest) (*BackupLocationResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResp := BackupLocationResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateAppRequest *publicbackuplocapi.UpdateBackupLocationRequest
	err = copier.Copy(&updateAppRequest, updateRequest)
	if err != nil {
		return nil, err
	}
	backupLocationModel, err := backupLocationClient.UpdateBackupLocation(ctx, updateAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	err = copier.Copy(&bckpLocResp, backupLocationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil

}

// SyncToBackupLocation returned synced backup location model.

// DeleteBackupLocation delete backup location and return status.
func (BackupLocGrpcV1 *PlatformGrpc) DeleteBackupLocation(backupLocation *BackupLocationRequest) error {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	deleteRequest := &publicbackuplocapi.DeleteBackupLocationRequest{Id: *backupLocation.List.Meta.Uid}
	_, err = backupLocationClient.DeleteBackupLocation(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error when calling `BackupLocationServiceDeleteBackupLocation`: %v\n", err)
	}
	return nil
}

func backupLocationConfig(request *BackupLocationRequest) *publicbackuplocapi.Config {
	createRequest := request.Create
	PROVIDER_TYPE := createRequest.Config.Provider.CloudProvider

	switch PROVIDER_TYPE {
	case PROVIDER_S3:
		log.Debugf("creating s3 credentials")
		return &publicbackuplocapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_S3),
			},
			CloudCredentialId: createRequest.Config.CloudCredentialsId,
			Location: &publicbackuplocapi.Config_S3Storage{
				S3Storage: &publicbackuplocapi.S3ObjectStorage{
					BucketName: createRequest.Config.BkpLocation.S3Storage.BucketName,
					Region:     createRequest.Config.BkpLocation.S3Storage.Region,
					Endpoint:   createRequest.Config.BkpLocation.S3Storage.Endpoint,
				},
			},
		}

	case PROVIDER_AZURE:
		log.Debugf("creating azure credentials")
		return &publicbackuplocapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_AZURE),
			},
			CloudCredentialId: createRequest.Config.CloudCredentialsId,
			Location: &publicbackuplocapi.Config_AzureStorage{
				AzureStorage: &publicbackuplocapi.AzureBlobStorage{
					ContainerName: createRequest.Config.BkpLocation.AzureStorage.ContainerName,
				},
			},
		}

	case PROVIDER_GOOGLE:
		log.Debugf("creating gcp credentials")
		return &publicbackuplocapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_GOOGLE),
			},
			CloudCredentialId: createRequest.Config.CloudCredentialsId,
			Location: &publicbackuplocapi.Config_GoogleStorage{
				GoogleStorage: &publicbackuplocapi.GoogleCloudStorage{
					BucketName: createRequest.Config.BkpLocation.GoogleStorage.BucketName,
				},
			},
		}

	default:
		log.Debugf("creating s3 credentials by default")
		return &publicbackuplocapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_S3),
			},
			CloudCredentialId: createRequest.Config.CloudCredentialsId,
			Location: &publicbackuplocapi.Config_S3Storage{
				S3Storage: &publicbackuplocapi.S3ObjectStorage{
					BucketName: createRequest.Config.BkpLocation.S3Storage.BucketName,
					Region:     createRequest.Config.BkpLocation.S3Storage.Region,
					Endpoint:   createRequest.Config.BkpLocation.S3Storage.Endpoint,
				},
			},
		}
	}
}

func copyCloudLocationResponse(bkpLocation *publicbackuplocapi.BackupLocation) *BackupLocationResponse {
	bkpLocResp := BackupLocationResponse{
		Create: BackupLocation{},
	}

	//Test Print
	log.Infof("bucket Name before copy [%s]", bkpLocation.Config.GetS3Storage().BucketName)
	log.Infof("end point before copy [%s]", bkpLocation.Config.GetS3Storage().Endpoint)
	log.Infof("region before copy [%s]", bkpLocation.Config.GetS3Storage().Region)

	switch bkpLocation.Config.Provider.GetCloudProvider() {
	case 3:
		log.Debugf("copying s3 location")
		bkpLocResp.Create.Meta.Uid = &bkpLocation.Meta.Uid
		bkpLocResp.Create.Meta.Name = &bkpLocation.Meta.Name
		bkpLocResp.Create.Config.CloudCredentialsId = bkpLocation.Config.GetCloudCredentialId()
		bkpLocResp.Create.Config.BkpLocation.S3Storage.BucketName = bkpLocation.Config.GetS3Storage().BucketName
		bkpLocResp.Create.Config.BkpLocation.S3Storage.Endpoint = bkpLocation.Config.GetS3Storage().Endpoint
		bkpLocResp.Create.Config.BkpLocation.S3Storage.Region = bkpLocation.Config.GetS3Storage().Region
		bkpLocResp.Create.Config.Provider.Name = "s3"
	case 1:
		log.Debugf("copying azure location")
		bkpLocResp.Create.Config.BkpLocation.AzureStorage.ContainerName = bkpLocation.Config.GetAzureStorage().ContainerName
		bkpLocResp.Create.Config.Provider.Name = "azure"
	case 2:
		log.Debugf("copying gcp credentials")
		bkpLocResp.Create.Config.BkpLocation.GoogleStorage.BucketName = bkpLocation.Config.GetGoogleStorage().BucketName
		bkpLocResp.Create.Config.Provider.Name = "gcp"
	}

	//Test Print
	log.Infof("bucket Name after copy [%s]", bkpLocResp.Create.Config.BkpLocation.S3Storage.BucketName)
	log.Infof("end point after copy [%s]", bkpLocResp.Create.Config.BkpLocation.S3Storage.Endpoint)
	log.Infof("region after copy [%s]", bkpLocResp.Create.Config.BkpLocation.S3Storage.Region)

	return &bkpLocResp

}