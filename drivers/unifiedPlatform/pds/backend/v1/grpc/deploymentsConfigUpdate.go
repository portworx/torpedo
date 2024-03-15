package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicdeploymentapis "github.com/pure-px/apis/public/portworx/pds/deployment/apiv1"
	publicdeploymentConfigUpdate "github.com/pure-px/apis/public/portworx/pds/deploymentconfigupdate/apiv1"
	deploymenttopology "github.com/pure-px/apis/public/portworx/pds/deploymenttopology/apiv1"
	"google.golang.org/grpc"
)

// GetClient updates the header with bearer token and returns the new client
func (deployment *PdsGrpc) getDeploymentConfigClient() (context.Context, publicdeploymentConfigUpdate.DeploymentConfigUpdateServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicdeploymentConfigUpdate.DeploymentConfigUpdateServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	depClient = publicdeploymentConfigUpdate.NewDeploymentConfigUpdateServiceClient(deployment.ApiClientV2)

	return ctx, depClient, token, nil
}

func (deployment *PdsGrpc) UpdateDeployment(updateDeploymentRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	depResponse := automationModels.WorkFlowResponse{}
	updateRequest := &publicdeploymentConfigUpdate.CreateDeploymentConfigUpdateRequest{
		DeploymentConfigUpdate: &publicdeploymentConfigUpdate.DeploymentConfigUpdate{
			Meta: &commonapiv1.Meta{
				Uid:             "",
				Name:            "",
				Description:     "",
				ResourceVersion: "",
				CreateTime:      nil,
				UpdateTime:      nil,
				Labels:          nil,
				Annotations:     nil,
				ParentReference: nil,
				ResourceNames:   nil,
			},
			Config: &publicdeploymentConfigUpdate.Config{
				DeploymentMeta: &commonapiv1.Meta{
					Uid:             "",
					Name:            "",
					Description:     "",
					ResourceVersion: "",
					CreateTime:      nil,
					UpdateTime:      nil,
					Labels:          nil,
					Annotations:     nil,
					ParentReference: nil,
					ResourceNames:   nil,
				},
				DeploymentConfig: &publicdeploymentapis.Config{
					References: nil,
					//TlsEnabled: false,
					DeploymentTopologies: []*deploymenttopology.DeploymentTopology{
						{
							Name:        "",
							Description: "",
							Replicas:    4,
							ResourceTemplate: &deploymenttopology.Template{
								Id:              "",
								ResourceVersion: "",
								Values:          nil,
							},
						},
					},
				},
			},
			Status: nil,
		},
	}

	ctx, client, _, err := deployment.getDeploymentConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while c: %v\n", err)
	}

	apiResponse, err := client.CreateDeploymentConfigUpdate(ctx, updateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while updating the deployment: %v\n", err)
	}
	copier.Copy(&depResponse, apiResponse)

	return &depResponse, nil
}
