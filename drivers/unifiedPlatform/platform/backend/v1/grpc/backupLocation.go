package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicbackuplocapi "github.com/pure-px/apis/public/portworx/platform/backuplocation/apiv1"
	publiccloudcredapi "github.com/pure-px/apis/public/portworx/platform/cloudcredential/apiv1"
	"google.golang.org/grpc"
)

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
func (BackupLocGrpcV1 *PlatformGrpc) ListBackupLocations(request *BackupLocation) ([]BackupLocation, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResponse := []BackupLocation{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	listbkpLocationRequest := &publicbackuplocapi.ListBackupLocationsRequest{
		TenantId:   request.TenantID,
		Pagination: NewPaginationRequest(1, 50),
	}
	backupLocationModels, err := backupLocationClient.ListBackupLocations(ctx, listbkpLocationRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("Value of tenants - [%v]", backupLocationModels)

	//err = copier.Copy(&bckpLocResponse, backupLocationModels.BackupLocations)
	//if err != nil {
	//	return nil, err
	//}

	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResponse)
	return bckpLocResponse, nil
}

// GetBackupLocation get backup location model by its ID.
func (BackupLocGrpcV1 *PlatformGrpc) GetBackupLocation(getReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	bckpLocResp := WorkFlowResponse{}
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

func backupLocationConfig(createRequest *BackupLocation) *publicbackuplocapi.Config {
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

func copyCloudLocationResponse(providerType int32, bkpLocation publicbackuplocapi.BackupLocation) *BackupLocation {
	bckpLocResp := BackupLocation{}

	//Test Print
	log.Infof("bucket Name before copy [%s]", bkpLocation.Config.GetS3Storage().BucketName)
	log.Infof("end point before copy [%s]", bkpLocation.Config.GetS3Storage().Endpoint)
	log.Infof("region before copy [%s]", bkpLocation.Config.GetS3Storage().Region)

	switch providerType {
	case PROVIDER_S3:
		log.Debugf("copying s3 location")
		bckpLocResp.Config.BkpLocation.S3Storage.BucketName = bkpLocation.Config.GetS3Storage().BucketName
		bckpLocResp.Config.BkpLocation.S3Storage.Endpoint = bkpLocation.Config.GetS3Storage().Endpoint
		bckpLocResp.Config.BkpLocation.S3Storage.Region = bkpLocation.Config.GetS3Storage().Region
	case PROVIDER_AZURE:
		log.Debugf("copying azure location")
		bckpLocResp.Config.BkpLocation.AzureStorage.ContainerName = bkpLocation.Config.GetAzureStorage().ContainerName
	case PROVIDER_GOOGLE:
		log.Debugf("copying gcp credentials")
		bckpLocResp.Config.BkpLocation.GoogleStorage.BucketName = bkpLocation.Config.GetGoogleStorage().BucketName
	}

	//Test Print
	log.Infof("bucket Name after copy [%s]", bckpLocResp.Config.BkpLocation.S3Storage.BucketName)
	log.Infof("end point after copy [%s]", bckpLocResp.Config.BkpLocation.S3Storage.Endpoint)
	log.Infof("region after copy [%s]", bckpLocResp.Config.BkpLocation.S3Storage.Region)

	return &bckpLocResp

}

// CreateBackupLocation return newly created backup location model.
func (BackupLocGrpcV1 *PlatformGrpc) CreateBackupLocation(createRequest *BackupLocation) (*BackupLocation, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	createAppRequest := &publicbackuplocapi.CreateBackupLocationRequest{
		TenantId: createRequest.TenantID,
		BackupLocation: &publicbackuplocapi.BackupLocation{
			Meta: &commonapiv1.Meta{
				Name: *createRequest.Meta.Name,
			},
			Config: backupLocationConfig(createRequest),
		},
	}

	backupLocationModel, err := backupLocationClient.CreateBackupLocation(ctx, createAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `BackupLocationServiceCreateBackupLocation` to create backup target - %v", err)
	}

	log.Infof("Value of backupLocation - [%v]", backupLocationModel)

	bkpLocationResponse := copyCloudLocationResponse(createRequest.Config.Provider.CloudProvider, *backupLocationModel)

	log.Infof("Value of backupLocation after copy - [%v]", bkpLocationResponse)
	return bkpLocationResponse, nil
}

// UpdateBackupLocation return updated backup location model.
func (BackupLocGrpcV1 *PlatformGrpc) UpdateBackupLocation(updateRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResp := WorkFlowResponse{}
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
func (BackupLocGrpcV1 *PlatformGrpc) DeleteBackupLocation(backupLocationID *WorkFlowRequest) error {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	deleteRequest := &publicbackuplocapi.DeleteBackupLocationRequest{Id: backupLocationID.Id}
	_, err = backupLocationClient.DeleteBackupLocation(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error when calling `BackupLocationServiceDeleteBackupLocation`: %v\n", err)
	}
	return nil
}
