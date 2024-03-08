package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
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
func (cloudCredGrpcV1 *PlatformGrpc) ListCloudCredentials() ([]WorkFlowResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	cloudCredsResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	firstPageRequest := &publiccloudcredapi.ListCloudCredentialsRequest{
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
	return cloudCredsResponse, nil
}

// GetCloudCredentials gets cloud credentials by ts id
func (cloudCredGrpcV1 *PlatformGrpc) GetCloudCredentials(cloudCredId *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	var getRequest *publiccloudcredapi.GetCloudCredentialRequest
	err = copier.Copy(&getRequest, cloudCredId)
	if err != nil {
		return nil, err
	}
	cloudCredModel, err := cloudCredsClient.GetCloudCredential(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `GetCloudCredential` call: %v\n", err)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	err = copier.Copy(&cloudCredsResponse, cloudCredModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredModel)
	return &cloudCredsResponse, nil
}

func cloudConfig(createRequest *WorkFlowRequest) *publiccloudcredapi.Config {
	PROVIDER_TYPE := createRequest.CloudCredentials.Config.Provider.CloudProvider
	secret := createRequest.CloudCredentials.Config.Credentials
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

func copyCloudCredResponse(providerType int32, cloudCredModel publiccloudcredapi.CloudCredential) (*WorkFlowResponse, error) {
	cloudCredResponse := WorkFlowResponse{}

	//Test Print
	log.Infof("access key [%s]", cloudCredModel.Config.GetS3Credentials().AccessKey)
	log.Infof("secret key [%s]", cloudCredModel.Config.GetS3Credentials().SecretKey)

	switch providerType {
	case PROVIDER_S3:
		log.Debugf("copying s3 credentials")
		cloudCredResponse.CloudConfig.Credentials.S3Credentials.AccessKey = cloudCredModel.Config.GetS3Credentials().AccessKey
		cloudCredResponse.CloudConfig.Credentials.S3Credentials.SecretKey = cloudCredModel.Config.GetS3Credentials().SecretKey
	case PROVIDER_AZURE:
		log.Debugf("copying azure credentials")
		cloudCredResponse.CloudConfig.Credentials.AzureCredentials.AccountKey = cloudCredModel.Config.GetAzureCredentials().StorageAccountKey
		cloudCredResponse.CloudConfig.Credentials.AzureCredentials.AccountName = cloudCredModel.Config.GetAzureCredentials().StorageAccountName
	case PROVIDER_GOOGLE:
		log.Debugf("copying gcp credentials")
		cloudCredResponse.CloudConfig.Credentials.GcpCredentials.ProjectId = cloudCredModel.Config.GetGoogleCredentials().ProjectId
		cloudCredResponse.CloudConfig.Credentials.GcpCredentials.Key = cloudCredModel.Config.GetGoogleCredentials().JsonKey
	}
	return &cloudCredResponse, nil
}

// CreateCloudCredentials return newly created cloud credentials
func (cloudCredGrpcV1 *PlatformGrpc) CreateCloudCredentials(createRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	createCloudCredRequest := &publiccloudcredapi.CreateCloudCredentialRequest{
		TenantId: createRequest.TenantId,
		CloudCredential: &publiccloudcredapi.CloudCredential{
			Meta: &commonapiv1.Meta{
				Uid:             "",
				Name:            *createRequest.CloudCredentials.Meta.Name,
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
	cloudCredResponse, err := copyCloudCredResponse(createRequest.CloudCredentials.Config.Provider.CloudProvider, *cloudCredModel)

	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredResponse)
	return cloudCredResponse, nil
}

// UpdateCloudCredentials return newly created cloud credentials
func (cloudCredGrpcV1 *PlatformGrpc) UpdateCloudCredentials(updateRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
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
func (cloudCredGrpcV1 *PlatformGrpc) DeleteCloudCredential(cloudCredId *WorkFlowRequest) error {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	deleteRequest := &publiccloudcredapi.DeleteCloudCredentialRequest{Id: cloudCredId.Id}
	_, err = cloudCredsClient.DeleteCloudCredential(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error when calling `DeleteCloudCredential`: %v\n", err)
	}
	return nil
}
