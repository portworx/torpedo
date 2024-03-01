package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/pkg/log"
	publicdeploymentapis "github.com/pure-px/apis/public/portworx/pds/deployment/apiv1"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"

	"google.golang.org/grpc"
)

type PdsGrpc struct {
	ApiClientV2 *grpc.ClientConn
	AccountId   string
}

// GetClient updates the header with bearer token and returns the new client
func (deployment *PdsGrpc) getDeploymentClient() (context.Context, publicdeploymentapis.DeploymentServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicdeploymentapis.DeploymentServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	depClient = publicdeploymentapis.NewDeploymentServiceClient(deployment.ApiClientV2)

	return ctx, depClient, token, nil
}

func (deployment *PdsGrpc) CreateDeployment(createDeploymentRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	depResponse := WorkFlowResponse{}
	createRequest := publicdeploymentapis.CreateDeploymentRequest{}

	ctx, client, _, err := deployment.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error while c: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)

	copier.Copy(&createRequest, createDeploymentRequest.Deployment)

	apiResponse, err := client.CreateDeployment(ctx, &createRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while getting the account: %v\n", err)
	}
	log.Infof("Value of accounts before copy - [%v]", apiResponse.Meta.Name)
	copier.Copy(&depResponse, apiResponse)
	log.Infof("Value of accounts after copy - [%v]", *depResponse.Meta.Name)

	return &depResponse, nil

}
