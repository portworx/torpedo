package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	cloudCredentialv1 "github.com/pure-px/platform-api-go-client/platform/v1/cloudcredential"
	status "net/http"
)

const (
	PROVIDER_UNSPECIFIED  int32 = 0
	PROVIDER_AZURE        int32 = 1
	PROVIDER_GOOGLE       int32 = 2
	PROVIDER_S3           int32 = 3
	PROVIDER_UNSTRUCTURED int32 = 4
)

// ListCloudCredentials return list of cloud credentials
func (cloudCred *PLATFORM_API_V1) ListCloudCredentials(request *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	cloudCredsResponse := CloudCredentialsResponse{
		List: ListCloudCredentials{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceListCloudCredentials(ctx).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceListcloudCredations`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	err = copier.Copy(&cloudCredsResponse, cloudCredModel.CloudCredentials)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// GetCloudCredentials gets cloud credentials by ts id
func (cloudCred *PLATFORM_API_V1) GetCloudCredentials(getReq *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceGetCloudCredential(ctx, getReq.Get.CloudCredentialsId).IncludeConfig(getReq.Get.IsConfigRequired).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceGetCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	cloudCredResponse := copyCloudCredResponse(getReq.Create.Config.Provider.CloudProvider, cloudCredModel)

	return cloudCredResponse, nil
}

// CreateCloudCredentials return newly created cloud credentials
func (cloudCred *PLATFORM_API_V1) CreateCloudCredentials(createRequest *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	ctx, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	v1CloudCred := cloudCredentialv1.V1CloudCredential{
		Meta: &cloudCredentialv1.V1Meta{
			Name: createRequest.Create.Meta.Name,
		},
		Config: cloudConfig(createRequest),
	}

	log.Debugf("After copy cloud cred [%+v]", v1CloudCred)
	log.Debugf("cloud provider [%s]", v1CloudCred.Config.Provider.GetCloudProvider())
	log.Debugf("cloud access key [%s]", v1CloudCred.Config.S3Credentials.GetConfigS3AccessKey())

	cloudCredModel, res, err := cloudCredsClient.CloudCredentialServiceCreateCloudCredential(ctx, createRequest.Create.TenantID).V1CloudCredential(v1CloudCred).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `CloudCredentialServiceCreateCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	cloudCredResponse := copyCloudCredResponse(createRequest.Create.Config.Provider.CloudProvider, cloudCredModel)

	return cloudCredResponse, nil
}

// UpdateCloudCredentials return updated created cloud credentials
func (cloudCred *PLATFORM_API_V1) UpdateCloudCredentials(updateReq *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	_, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	cloudCredsResponse := CloudCredentialsResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateAppReq cloudCredentialv1.ApiCloudCredentialServiceUpdateCloudCredentialRequest
	err = copier.Copy(&updateAppReq, updateReq)
	if err != nil {
		return nil, err
	}
	cloudCredationModel, res, err := cloudCredsClient.CloudCredentialServiceUpdateCloudCredentialExecute(updateAppReq)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `cloudCredationServiceUpdatecloudCredation`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&cloudCredsResponse, cloudCredationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// DeleteCloudCredential delete cloud cred model.
func (cloudCred *PLATFORM_API_V1) DeleteCloudCredential(cloudCreds *CloudCredentialsRequest) error {
	ctx, cloudCredsClient, err := cloudCred.getCloudCredentialClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := cloudCredsClient.CloudCredentialServiceDeleteCloudCredential(ctx, cloudCreds.Get.CloudCredentialsId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `CloudCredentialServiceDeleteCloudCredential`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}

func cloudConfig(createRequest *CloudCredentialsRequest) *cloudCredentialv1.V1Config {
	PROVIDER_TYPE := createRequest.Create.Config.Provider.CloudProvider
	secret := createRequest.Create.Config
	switch PROVIDER_TYPE {
	case PROVIDER_S3:
		log.Debugf("creating s3 credentials")
		return &cloudCredentialv1.V1Config{
			Provider: &cloudCredentialv1.V1Provider{
				CloudProvider: cloudCredentialv1.V1ProviderType.Ptr("S3COMPATIBLE"),
			},
			S3Credentials: &cloudCredentialv1.V1S3Credentials{
				ConfigS3AccessKey: &secret.S3Credentials.AccessKey,
				ConfigS3SecretKey: &secret.S3Credentials.SecretKey,
			},
		}
	case PROVIDER_AZURE:
		log.Debugf("creating azure credentials")
		return &cloudCredentialv1.V1Config{
			Provider: &cloudCredentialv1.V1Provider{
				CloudProvider: cloudCredentialv1.V1ProviderType.Ptr("AZURE"),
			},
			AzureCredentials: &cloudCredentialv1.V1AzureCredentials{
				ConfigAzureStorageAccountKey:  &secret.S3Credentials.AccessKey,
				ConfigAzureStorageAccountName: &secret.S3Credentials.SecretKey,
			},
		}
	case PROVIDER_GOOGLE:
		log.Debugf("creating gcp credentials")
		return &cloudCredentialv1.V1Config{
			Provider: &cloudCredentialv1.V1Provider{
				CloudProvider: cloudCredentialv1.V1ProviderType.Ptr("GOOGLE"),
			},
			GoogleCredentials: &cloudCredentialv1.V1GoogleCredentials{
				ConfigGoogleJsonKey:   &secret.S3Credentials.AccessKey,
				ConfigGoogleProjectId: &secret.S3Credentials.SecretKey,
			},
		}

	default:
		log.Debugf("creating s3 credentials by default")
		return &cloudCredentialv1.V1Config{
			Provider: &cloudCredentialv1.V1Provider{
				CloudProvider: cloudCredentialv1.V1ProviderType.Ptr("S3COMPATIBLE"),
			},
			S3Credentials: &cloudCredentialv1.V1S3Credentials{
				ConfigS3AccessKey: &secret.S3Credentials.AccessKey,
				ConfigS3SecretKey: &secret.S3Credentials.SecretKey,
			},
		}
	}
}

func copyCloudCredResponse(providerType int32, cloudCredModel *cloudCredentialv1.V1CloudCredential) *CloudCredentialsResponse {
	cloudCredResponse := CloudCredentialsResponse{}

	//Test Print
	log.Infof("access key before copy [%s]", *cloudCredModel.Config.GetS3Credentials().ConfigS3AccessKey)
	log.Infof("secret key before copy [%s]", *cloudCredModel.Config.GetS3Credentials().ConfigS3SecretKey)

	switch providerType {
	case PROVIDER_S3:
		log.Debugf("copying s3 credentials")
		cloudCredResponse.Create.Config.S3Credentials.AccessKey = *cloudCredModel.Config.GetS3Credentials().ConfigS3AccessKey
		cloudCredResponse.Create.Config.S3Credentials.SecretKey = *cloudCredModel.Config.GetS3Credentials().ConfigS3SecretKey
		cloudCredResponse.Create.Meta.Uid = cloudCredModel.Meta.Uid
		cloudCredResponse.Create.Meta.Name = cloudCredModel.Meta.Name

	case PROVIDER_AZURE:
		log.Debugf("copying azure credentials")
		cloudCredResponse.Create.Config.AzureCredentials.AccountKey = *cloudCredModel.Config.GetAzureCredentials().ConfigAzureStorageAccountKey
		cloudCredResponse.Create.Config.AzureCredentials.AccountName = *cloudCredModel.Config.GetAzureCredentials().ConfigAzureStorageAccountName
		cloudCredResponse.Create.Meta.Uid = cloudCredModel.Meta.Uid
		cloudCredResponse.Create.Meta.Name = cloudCredModel.Meta.Name

	case PROVIDER_GOOGLE:
		log.Debugf("copying gcp credentials")
		cloudCredResponse.Create.Config.GoogleCredentials.ProjectId = *cloudCredModel.Config.GetGoogleCredentials().ConfigGoogleProjectId
		cloudCredResponse.Create.Config.GoogleCredentials.Key = *cloudCredModel.Config.GetGoogleCredentials().ConfigGoogleJsonKey
		cloudCredResponse.Create.Meta.Uid = cloudCredModel.Meta.Uid
		cloudCredResponse.Create.Meta.Name = cloudCredModel.Meta.Name
	}

	//Test Print
	log.Infof("access key after copy [%s]", cloudCredResponse.Create.Config.S3Credentials.AccessKey)
	log.Infof("secret key after copy [%s]", cloudCredResponse.Create.Config.S3Credentials.SecretKey)

	return &cloudCredResponse
}
