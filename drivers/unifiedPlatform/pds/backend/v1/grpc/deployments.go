package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicdeploymentapis "github.com/pure-px/apis/public/portworx/pds/deployment/apiv1"
	deploymenttopology "github.com/pure-px/apis/public/portworx/pds/deploymenttopology/apiv1"

	"google.golang.org/grpc"
)

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
	dep := createDeploymentRequest.Deployment.V1Deployment

	createRequest := &publicdeploymentapis.CreateDeploymentRequest{
		NamespaceId: createDeploymentRequest.Deployment.NamespaceID,
		Deployment: &publicdeploymentapis.Deployment{
			Meta: nil,
			Config: &publicdeploymentapis.Config{
				References: nil,
				TlsEnabled: false,
				DeploymentTopologies: []*deploymenttopology.DeploymentTopology{
					{
						Name:        *dep.Meta.Name,
						Description: "",
						Replicas:    3,
						ResourceTemplate: &deploymenttopology.Template{
							Id:              *dep.Config.DeploymentTopologies[0].ResourceTemplate.Id,
							ResourceVersion: "",
							Values:          nil,
						},
					},
				},
			},
		},
	}

	ctx, client, _, err := deployment.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	log.Debugf("Account ID: [%s]", deployment.AccountId)
	log.Debugf("Namespace ID: [%s]", createDeploymentRequest.Deployment.NamespaceID)
	log.Debugf("workflowrequest ResourceTemplateId: [%s]", *createDeploymentRequest.Deployment.V1Deployment.Config.DeploymentTopologies[0].ResourceTemplate.Id)

	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)

	apiResponse, err := client.CreateDeployment(ctx, createRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while creating the deployment: %v\n", err)
	}

	err = copier.Copy(&depResponse, apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while copying the response:%v\n", err)
	}

	log.Infof("Value of response app template after copy - [%v]", depResponse)

	return &depResponse, nil

}
