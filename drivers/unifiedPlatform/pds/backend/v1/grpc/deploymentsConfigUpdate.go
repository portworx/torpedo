package grpc

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicdeploymentapis "github.com/pure-px/apis/public/portworx/pds/dataservicedeployment/apiv1"
	publicdeploymentConfigUpdate "github.com/pure-px/apis/public/portworx/pds/dataservicedeploymentconfigupdate/apiv1"
	deploymenttopology "github.com/pure-px/apis/public/portworx/pds/dataservicedeploymenttopology/apiv1"
	"google.golang.org/grpc"
)

// GetClient updates the header with bearer token and returns the new client
func (deployment *PdsGrpc) getDeploymentConfigClient() (context.Context, publicdeploymentConfigUpdate.DataServiceDeploymentConfigUpdateServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicdeploymentConfigUpdate.DataServiceDeploymentConfigUpdateServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	depClient = publicdeploymentConfigUpdate.NewDataServiceDeploymentConfigUpdateServiceClient(deployment.ApiClientV2)

	return ctx, depClient, token, nil
}

func (deployment *PdsGrpc) GetDeploymentConfig(updateDeploymentRequest *automationModels.PDSDeploymentRequest) (*automationModels.PDSDeploymentResponse, error) {
	depResponse := automationModels.PDSDeploymentResponse{
		Update: automationModels.V1DeploymentUpdate{},
	}

	//TODO: Fill this once we take up grpc client validation

	return &depResponse, nil
}

func (deployment *PdsGrpc) UpdateDeployment(updateDeploymentRequest *automationModels.PDSDeploymentRequest) (*automationModels.PDSDeploymentResponse, error) {
	depResponse := automationModels.PDSDeploymentResponse{
		Update: automationModels.V1DeploymentUpdate{},
	}

	//TODO: try copy else go with the below approach
	updateRequest := &publicdeploymentConfigUpdate.CreateDataServiceDeploymentConfigUpdateRequest{
		DataServiceDeploymentConfigUpdate: &publicdeploymentConfigUpdate.DataServiceDeploymentConfigUpdate{
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
				DataServiceDeploymentMeta: &commonapiv1.Meta{
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
				DataServiceDeploymentConfig: &publicdeploymentapis.Config{
					References: nil,
					//TlsEnabled: false,
					DataServiceDeploymentTopologies: []*deploymenttopology.DataServiceDeploymentTopology{
						{
							Name:        "",
							Description: "",
							Instances:   4,
							ResourceSettings: &deploymenttopology.Template{
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

	apiResponse, err := client.CreateDataServiceDeploymentConfigUpdate(ctx, updateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while updating the deployment: %v\n", err)
	}
	err = utilities.CopyStruct(apiResponse, &depResponse.Update)
	if err != nil {
		return nil, fmt.Errorf("Error while copying the updated dep response : %v\n", err)
	}

	return &depResponse, nil
}
