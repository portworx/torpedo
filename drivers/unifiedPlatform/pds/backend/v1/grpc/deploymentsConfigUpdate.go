package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicdeploymentapis "github.com/pure-px/apis/public/portworx/pds/deploymentconfigupdate/apiv1"
	"google.golang.org/grpc"
)

// GetClient updates the header with bearer token and returns the new client
func (deployment *PdsGrpc) getDeploymentConfigClient() (context.Context, publicdeploymentapis.DeploymentConfigUpdateServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicdeploymentapis.DeploymentConfigUpdateServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	depClient = publicdeploymentapis.NewDeploymentConfigUpdateServiceClient(deployment.ApiClientV2)

	return ctx, depClient, token, nil
}

func (deployment *PdsGrpc) UpdateDeploymentConfig(updateDeploymentRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	depResponse := apiStructs.WorkFlowResponse{}
	updateRequest := publicdeploymentapis.CreateDeploymentConfigUpdateRequest{}

	ctx, client, _, err := deployment.getDeploymentConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while c: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)

	copier.Copy(&updateRequest, updateDeploymentRequest)

	apiResponse, err := client.CreateDeploymentConfigUpdate(ctx, &updateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while getting the account: %v\n", err)
	}
	copier.Copy(&depResponse, apiResponse)

	return &depResponse, nil

}
