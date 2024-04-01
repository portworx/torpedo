package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publiccloudcredapi "github.com/pure-px/apis/public/portworx/platform/cloudcredential/apiv1"
	"google.golang.org/grpc"
)

const (
	PROVIDER_UNSPECIFIED  int32 = 0
	PROVIDER_AZURE        int32 = 1
	PROVIDER_GOOGLE       int32 = 2
	PROVIDER_S3           int32 = 3
	PROVIDER_UNSTRUCTURED int32 = 4
)

// getCloudCredClient updates the header with bearer token and returns the new client
func (cloudCredGrpcV1 *PlatformGrpc) getCloudCredClient() (context.Context, publiccloudcredapi.CloudCredentialServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var backupLocClient publiccloudcredapi.CloudCredentialServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	ctx = WithAccountIDMetaCtx(ctx, cloudCredGrpcV1.AccountId)

	backupLocClient = publiccloudcredapi.NewCloudCredentialServiceClient(cloudCredGrpcV1.ApiClientV1)

	return ctx, backupLocClient, token, nil
}

// ListCloudCredentials return list of cloud credentials
func (cloudCredGrpcV1 *PlatformGrpc) ListCloudCredentials(request *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	cloudCredsResponse := CloudCredentialsResponse{
		List: ListCloudCredentials{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	firstPageRequest := &publiccloudcredapi.ListCloudCredentialsRequest{
		TenantId:   request.Create.TenantID,
		Pagination: NewPaginationRequest(1, 50),
	}
	cloudCredModel, err := cloudCredsClient.ListCloudCredentials(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `ListCloudCredentials` call: %v\n", err)
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
func (cloudCredGrpcV1 *PlatformGrpc) GetCloudCredentials(getWorkflowRequest *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	getRequest := &publiccloudcredapi.GetCloudCredentialRequest{
		Id:            getWorkflowRequest.Get.CloudCredentialsId,
		IncludeConfig: getWorkflowRequest.Get.IsConfigRequired,
	}

	cloudCredModel, err := cloudCredsClient.GetCloudCredential(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `GetCloudCredential` call: %v\n", err)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)

	cloudCredResponse := copyCloudCredResponse(getWorkflowRequest.Create.Config.Provider.CloudProvider, cloudCredModel)

	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredResponse)
	return cloudCredResponse, nil
}

func cloudConfig(createRequest *CloudCredentialsRequest) *publiccloudcredapi.Config {
	PROVIDER_TYPE := createRequest.Create.Config.Provider.CloudProvider
	secret := createRequest.Create.Config.Credentials
	switch PROVIDER_TYPE {
	case PROVIDER_S3:
		log.Debugf("creating s3 credentials")
		return &publiccloudcredapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_S3),
			},
			Credentials: &publiccloudcredapi.Config_S3Credentials{
				S3Credentials: &publiccloudcredapi.S3Credentials{
					AccessKey: secret.S3Credentials.AccessKey,
					SecretKey: secret.S3Credentials.SecretKey,
				},
			},
		}
	case PROVIDER_AZURE:
		log.Debugf("creating azure credentials")
		return &publiccloudcredapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_AZURE),
			},
			Credentials: &publiccloudcredapi.Config_AzureCredentials{
				AzureCredentials: &publiccloudcredapi.AzureCredentials{
					StorageAccountName: secret.AzureCredentials.AccountName,
					StorageAccountKey:  secret.AzureCredentials.AccountKey,
				},
			},
		}
	case PROVIDER_GOOGLE:
		log.Debugf("creating gcp credentials")
		return &publiccloudcredapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_GOOGLE),
			},
			Credentials: &publiccloudcredapi.Config_GoogleCredentials{
				GoogleCredentials: &publiccloudcredapi.GoogleCredentials{
					ProjectId: secret.GcpCredentials.ProjectId,
					JsonKey:   secret.GcpCredentials.Key,
				},
			},
		}

	default:
		log.Debugf("creating s3 credentials by default")
		return &publiccloudcredapi.Config{
			Provider: &publiccloudcredapi.Provider{
				CloudProvider: publiccloudcredapi.Provider_Type(PROVIDER_S3),
			},
			Credentials: &publiccloudcredapi.Config_S3Credentials{
				S3Credentials: &publiccloudcredapi.S3Credentials{
					AccessKey: secret.S3Credentials.AccessKey,
					SecretKey: secret.S3Credentials.SecretKey,
				},
			},
		}
	}
}

func copyCloudCredResponse(providerType int32, cloudCredModel *publiccloudcredapi.CloudCredential) *CloudCredentialsResponse {
	cloudCredResponse := CloudCredentialsResponse{}

	//Test Print
	log.Infof("access key before copy [%s]", cloudCredModel.Config.GetS3Credentials().AccessKey)
	log.Infof("secret key before copy [%s]", cloudCredModel.Config.GetS3Credentials().SecretKey)

	switch providerType {
	case PROVIDER_S3:
		log.Debugf("copying s3 credentials")
		cloudCredResponse.Create.Config.Credentials.S3Credentials.AccessKey = cloudCredModel.Config.GetS3Credentials().AccessKey
		cloudCredResponse.Create.Config.Credentials.S3Credentials.SecretKey = cloudCredModel.Config.GetS3Credentials().SecretKey
		cloudCredResponse.Create.Meta.Uid = &cloudCredModel.Meta.Uid
		cloudCredResponse.Create.Meta.Name = &cloudCredModel.Meta.Name

	case PROVIDER_AZURE:
		log.Debugf("copying azure credentials")
		cloudCredResponse.Create.Config.Credentials.AzureCredentials.AccountKey = cloudCredModel.Config.GetAzureCredentials().StorageAccountKey
		cloudCredResponse.Create.Config.Credentials.AzureCredentials.AccountName = cloudCredModel.Config.GetAzureCredentials().StorageAccountName
		cloudCredResponse.Create.Meta.Uid = &cloudCredModel.Meta.Uid
		cloudCredResponse.Create.Meta.Name = &cloudCredModel.Meta.Name

	case PROVIDER_GOOGLE:
		log.Debugf("copying gcp credentials")
		cloudCredResponse.Create.Config.Credentials.GcpCredentials.ProjectId = cloudCredModel.Config.GetGoogleCredentials().ProjectId
		cloudCredResponse.Create.Config.Credentials.GcpCredentials.Key = cloudCredModel.Config.GetGoogleCredentials().JsonKey
		cloudCredResponse.Create.Meta.Uid = &cloudCredModel.Meta.Uid
		cloudCredResponse.Create.Meta.Name = &cloudCredModel.Meta.Name
	}

	//Test Print
	log.Infof("access key after copy [%s]", cloudCredResponse.Create.Config.Credentials.S3Credentials.AccessKey)
	log.Infof("secret key after copy [%s]", cloudCredResponse.Create.Config.Credentials.S3Credentials.SecretKey)

	return &cloudCredResponse
}

// CreateCloudCredentials return newly created cloud credentials
func (cloudCredGrpcV1 *PlatformGrpc) CreateCloudCredentials(createRequest *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	createCloudCredRequest := &publiccloudcredapi.CreateCloudCredentialRequest{
		TenantId: createRequest.Create.TenantID,
		CloudCredential: &publiccloudcredapi.CloudCredential{
			Meta: &commonapiv1.Meta{
				Uid:             "",
				Name:            *createRequest.Create.Meta.Name,
				Description:     "",
				ResourceVersion: "3",
				CreateTime:      nil,
				UpdateTime:      nil,
				Labels:          nil,
				Annotations:     nil,
				ParentReference: nil,
				ResourceNames:   nil,
			},
			Config: cloudConfig(createRequest),
		},
	}

	ctx = WithAccountIDMetaCtx(ctx, cloudCredGrpcV1.AccountId)

	cloudCredModel, err := cloudCredsClient.CreateCloudCredential(ctx, createCloudCredRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `CloudCredentialServiceCreateCloudCredential` to create cloud credential - %v", err)
	}

	log.Infof("cloud cred response [%+v]", cloudCredModel)
	cloudCredResponse := copyCloudCredResponse(createRequest.Create.Config.Provider.CloudProvider, cloudCredModel)

	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredResponse)
	return cloudCredResponse, nil
}

// UpdateCloudCredentials return newly created cloud credentials
func (cloudCredGrpcV1 *PlatformGrpc) UpdateCloudCredentials(updateRequest *CloudCredentialsRequest) (*CloudCredentialsResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := CloudCredentialsResponse{
		Update: CloudCredentials{},
	}
	var updateAppRequest *publiccloudcredapi.UpdateCloudCredentialRequest
	err = copier.Copy(&updateAppRequest, updateRequest)
	if err != nil {
		return nil, err
	}
	cloudCredModel, err := cloudCredsClient.UpdateCloudCredential(ctx, updateAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `CloudCredentialServiceCreateCloudCredential` to create cloud credential - %v", err)
	}
	err = copier.Copy(&cloudCredsResponse, cloudCredModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// DeleteCloudCredential delete cloud cred model.
func (cloudCredGrpcV1 *PlatformGrpc) DeleteCloudCredential(cloudCred *CloudCredentialsRequest) error {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	deleteRequest := &publiccloudcredapi.DeleteCloudCredentialRequest{Id: cloudCred.Get.CloudCredentialsId}
	_, err = cloudCredsClient.DeleteCloudCredential(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error when calling `DeleteCloudCredential`: %v\n", err)
	}
	return nil
}
