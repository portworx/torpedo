package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publiccloudcredapi "github.com/pure-px/apis/public/portworx/platform/cloudcredential/apiv1"
	"google.golang.org/grpc"
)

type CloudCredentialGrpc struct {
	ApiClientV1 *grpc.ClientConn
}

// getCloudCredClient updates the header with bearer token and returns the new client
func (cloudCredGrpcV1 *CloudCredentialGrpc) getCloudCredClient() (context.Context, publiccloudcredapi.CloudCredentialServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var backupLocClient publiccloudcredapi.CloudCredentialServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	backupLocClient = publiccloudcredapi.NewCloudCredentialServiceClient(cloudCredGrpcV1.ApiClientV1)

	return ctx, backupLocClient, token, nil
}

// ListCloudCredentials return list of cloud credentials
func (cloudCredGrpcV1 *CloudCredentialGrpc) ListCloudCredentials() ([]WorkFlowResponse, error) {
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
	copier.Copy(&cloudCredsResponse, cloudCredModel.CloudCredentials)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return cloudCredsResponse, nil
}

// GetCloudCredentials gets cloud credentials by ts id
func (cloudCredGrpcV1 *CloudCredentialGrpc) GetCloudCredentials(cloudCredId *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	var getRequest *publiccloudcredapi.GetCloudCredentialRequest
	copier.Copy(&getRequest, cloudCredId)
	cloudCredModel, err := cloudCredsClient.GetCloudCredential(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `GetCloudCredential` call: %v\n", err)
	}
	log.Infof("Value of cloudCredentials - [%v]", cloudCredModel)
	copier.Copy(&cloudCredsResponse, cloudCredModel)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredModel)
	return &cloudCredsResponse, nil
}

// CreateCloudCredentials return newly created cloud credentials
func (cloudCredGrpcV1 *CloudCredentialGrpc) CreateCloudCredentials(createRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	var createCloudCredRequest *publiccloudcredapi.CreateCloudCredentialRequest
	copier.Copy(&createCloudCredRequest, createRequest)
	cloudCredModel, err := cloudCredsClient.CreateCloudCredential(ctx, createCloudCredRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `CloudCredentialServiceCreateCloudCredential` to create cloud credential - %v", err)
	}
	copier.Copy(&cloudCredsResponse, cloudCredModel)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// UpdateCloudCredentials return newly created cloud credentials
func (cloudCredGrpcV1 *CloudCredentialGrpc) UpdateCloudCredentials(updateRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, cloudCredsClient, _, err := cloudCredGrpcV1.getCloudCredClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	cloudCredsResponse := WorkFlowResponse{}
	var updateAppRequest *publiccloudcredapi.UpdateCloudCredentialRequest
	copier.Copy(&updateAppRequest, updateRequest)
	cloudCredModel, err := cloudCredsClient.UpdateCloudCredential(ctx, updateAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `CloudCredentialServiceCreateCloudCredential` to create cloud credential - %v", err)
	}
	copier.Copy(&cloudCredsResponse, cloudCredModel)
	log.Infof("Value of cloudCredentials after copy - [%v]", cloudCredsResponse)
	return &cloudCredsResponse, nil
}

// DeleteCloudCredential delete cloud cred model.
func (cloudCredGrpcV1 *CloudCredentialGrpc) DeleteCloudCredential(cloudCredId *WorkFlowRequest) error {
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
