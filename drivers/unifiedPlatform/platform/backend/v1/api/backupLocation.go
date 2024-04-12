package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/pure-px/platform-api-go-client/platform/v1/backuplocation"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	status "net/http"
)

// ListBackupLocations return lis of backup locatiobackuploc
func (backuploc *PLATFORM_API_V1) ListBackupLocations(request *BackupLocationRequest) (*BackupLocationResponse, error) {
	ctx, backupLocationClient, err := backuploc.getBackupLocClient()
	backupLocResp := BackupLocationResponse{
		List: ListBackupLocation{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	backupLocationModels, res, err := backupLocationClient.BackupLocationServiceListBackupLocations(ctx).TenantId(request.List.TenantID).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocatiobackuplocerviceListBackupLocatiobackuploc`: %v\n.Full HTTP respobackuploce: %v", err, res)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModels)

	for _, bkpLocation := range backupLocationModels.BackupLocations {
		log.Infof("printing cloud provider type [%s]", bkpLocation.Config.Provider.GetCloudProvider())
		resp := copyCloudLocationResponse(&bkpLocation)
		backupLocResp.List.BackupLocations = append(backupLocResp.List.BackupLocations, resp.Create)
	}

	log.Infof("Value of backupLocation after copy - [%v]", backupLocResp)
	return &backupLocResp, nil
}

// GetBackupLocation get backup location model by its ID.
func (backuploc *PLATFORM_API_V1) GetBackupLocation(getReq *WorkFlowRequest) (*BackupLocationResponse, error) {
	_, backupLocationClient, err := backuploc.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	bckpLocResp := BackupLocationResponse{}
	var getRequest backuplocation.ApiBackupLocationServiceGetBackupLocationRequest
	err = copier.Copy(&getRequest, getReq)
	if err != nil {
		return nil, err
	}
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceGetBackupLocationExecute(getRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when called `BackupLocatiobackuplocerviceGetBackupLocation`, Full HTTP respobackuploce: %v\n", res)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModel)
	err = copier.Copy(&bckpLocResp, backupLocationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

func backupLocationConfig(request *BackupLocationRequest) *backuplocation.Platformbackuplocationv1Config {
	createRequest := request.Create
	PROVIDER_TYPE := createRequest.Config.Provider.CloudProvider

	switch PROVIDER_TYPE {
	case PROVIDER_S3:
		log.Debugf("creating s3 bkpLocation")
		return &backuplocation.Platformbackuplocationv1Config{
			Provider: &backuplocation.V1Provider{
				CloudProvider: backuplocation.V1ProviderType.Ptr("S3COMPATIBLE"),
			},
			CloudCredentialId: &createRequest.Config.CloudCredentialsId,
			S3Storage: &backuplocation.V1S3ObjectStorage{
				BucketName: &createRequest.Config.BkpLocation.S3Storage.BucketName,
				Region:     &createRequest.Config.BkpLocation.S3Storage.Region,
				Endpoint:   &createRequest.Config.BkpLocation.S3Storage.Endpoint,
			},
		}

	case PROVIDER_AZURE:
		log.Debugf("creating azure bkpLocation")
		return &backuplocation.Platformbackuplocationv1Config{
			Provider: &backuplocation.V1Provider{
				CloudProvider: backuplocation.V1ProviderType.Ptr("AZURE"),
			},
			CloudCredentialId: &createRequest.Config.CloudCredentialsId,
			AzureStorage: &backuplocation.V1AzureBlobStorage{
				ContainerName: &createRequest.Config.BkpLocation.AzureStorage.ContainerName,
			},
		}

	case PROVIDER_GOOGLE:
		log.Debugf("creating gcp bkpLocation")
		return &backuplocation.Platformbackuplocationv1Config{
			Provider: &backuplocation.V1Provider{
				CloudProvider: backuplocation.V1ProviderType.Ptr("GOOGLE"),
			},
			CloudCredentialId: &createRequest.Config.CloudCredentialsId,
			GoogleStorage: &backuplocation.V1GoogleCloudStorage{
				BucketName: &createRequest.Config.BkpLocation.GoogleStorage.BucketName,
			},
		}

	default:
		log.Debugf("creating s3 bkpLocation by default")
		return &backuplocation.Platformbackuplocationv1Config{
			Provider: &backuplocation.V1Provider{
				CloudProvider: backuplocation.V1ProviderType.Ptr("S3COMPATIBLE"),
			},
			CloudCredentialId: &createRequest.Config.CloudCredentialsId,
			S3Storage: &backuplocation.V1S3ObjectStorage{
				BucketName: &createRequest.Config.BkpLocation.S3Storage.BucketName,
				Region:     &createRequest.Config.BkpLocation.S3Storage.Region,
				Endpoint:   &createRequest.Config.BkpLocation.S3Storage.Endpoint,
			},
		}
	}
}

func copyCloudLocationResponse(bkpLocation *backuplocation.V1BackupLocation) *BackupLocationResponse {
	bkpLocResp := BackupLocationResponse{
		Create: BackupLocation{},
	}

	switch bkpLocation.Config.Provider.GetCloudProvider() {
	case "S3COMPATIBLE":
		log.Debugf("copying s3 location")
		//Test Print
		log.Infof("bucket Name before copy [%s]", bkpLocation.Config.GetS3Storage().BucketName)
		log.Infof("end point before copy [%s]", bkpLocation.Config.GetS3Storage().Endpoint)
		log.Infof("region before copy [%s]", bkpLocation.Config.GetS3Storage().Region)
		log.Infof("bkp location cloud provider [%s]", bkpLocation.Config.Provider.GetCloudProvider())

		bkpLocResp.Create.Meta.Uid = bkpLocation.Meta.Uid
		bkpLocResp.Create.Meta.Name = bkpLocation.Meta.Name
		bkpLocResp.Create.Config.CloudCredentialsId = bkpLocation.Config.GetCloudCredentialId()
		bkpLocResp.Create.Config.BkpLocation.S3Storage.BucketName = *bkpLocation.Config.GetS3Storage().BucketName
		bkpLocResp.Create.Config.BkpLocation.S3Storage.Endpoint = *bkpLocation.Config.GetS3Storage().Endpoint
		bkpLocResp.Create.Config.BkpLocation.S3Storage.Region = *bkpLocation.Config.GetS3Storage().Region
		bkpLocResp.Create.Config.Provider.Name = "s3"
		//Test Print
		log.Infof("bucket Name after copy [%s]", bkpLocResp.Create.Config.BkpLocation.S3Storage.BucketName)
		log.Infof("end point after copy [%s]", bkpLocResp.Create.Config.BkpLocation.S3Storage.Endpoint)
		log.Infof("region after copy [%s]", bkpLocResp.Create.Config.BkpLocation.S3Storage.Region)
	case "AZURE":
		log.Debugf("copying azure location")
		bkpLocResp.Create.Meta.Uid = bkpLocation.Meta.Uid
		bkpLocResp.Create.Meta.Name = bkpLocation.Meta.Name
		bkpLocResp.Create.Config.BkpLocation.AzureStorage.ContainerName = *bkpLocation.Config.GetAzureStorage().ContainerName
		bkpLocResp.Create.Config.CloudCredentialsId = bkpLocation.Config.GetCloudCredentialId()
		bkpLocResp.Create.Config.Provider.Name = "azure"
	case "GOOGLE":
		log.Debugf("copying gcp credentials")
		bkpLocResp.Create.Meta.Uid = bkpLocation.Meta.Uid
		bkpLocResp.Create.Meta.Name = bkpLocation.Meta.Name
		bkpLocResp.Create.Config.BkpLocation.GoogleStorage.BucketName = *bkpLocation.Config.GetGoogleStorage().BucketName
		bkpLocResp.Create.Config.CloudCredentialsId = bkpLocation.Config.GetCloudCredentialId()
		bkpLocResp.Create.Config.Provider.Name = "gcp"
	}

	return &bkpLocResp

}

// CreateBackupLocation return newly created backup location model.
func (backuploc *PLATFORM_API_V1) CreateBackupLocation(createReq *BackupLocationRequest) (*BackupLocationResponse, error) {
	ctx, backupLocationClient, err := backuploc.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	createBackupLoc := &backuplocation.V1BackupLocation{
		Meta: &backuplocation.V1Meta{
			Name: createReq.Create.Meta.Name,
		},
		Config: backupLocationConfig(createReq),
	}

	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceCreateBackupLocation(ctx, createReq.Create.TenantID).V1BackupLocation(*createBackupLoc).Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocationServiceCreateBackupLocation`: %v\n.Full HTTP response: %v", err, res)
	}

	bkpLocationResponse := copyCloudLocationResponse(backupLocationModel)

	return bkpLocationResponse, nil
}

// UpdateBackupLocation return updated backup location model.
func (backuploc *PLATFORM_API_V1) UpdateBackupLocation(updateReq *WorkFlowRequest) (*BackupLocationResponse, error) {
	_, backupLocationClient, err := backuploc.getBackupLocClient()
	bckpLocResp := BackupLocationResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateBackupLoc backuplocation.ApiBackupLocationServiceUpdateBackupLocationRequest
	err = copier.Copy(&updateBackupLoc, updateReq)
	if err != nil {
		return nil, err
	}
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceUpdateBackupLocationExecute(updateBackupLoc)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocatiobackuplocerviceUpdateBackupLocation`: %v\n.Full HTTP respobackuploce: %v", err, res)
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
func (backuploc *PLATFORM_API_V1) DeleteBackupLocation(backupLocation *BackupLocationRequest) error {
	ctx, backupLocationClient, err := backuploc.getBackupLocClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := backupLocationClient.BackupLocationServiceDeleteBackupLocation(ctx, *backupLocation.List.Meta.Uid).Execute()
	if err != nil {
		return fmt.Errorf("Error when calling `BackupLocatiobackuplocerviceDeleteBackupLocation`: %v\n.Full HTTP respobackuploce: %v", err, res)
	}
	return nil
}
