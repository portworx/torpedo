package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicdeploymentapis "github.com/pure-px/apis/public/portworx/pds/dataservicedeployment/apiv1"
	deploymenttopology "github.com/pure-px/apis/public/portworx/pds/dataservicedeploymenttopology/apiv1"
	"google.golang.org/grpc"
)

// getDeploymentClient updates the header with bearer token and returns the new client
func (deployment *PdsGrpc) getDeploymentClient() (context.Context, publicdeploymentapis.DataServiceDeploymentServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicdeploymentapis.DataServiceDeploymentServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	depClient = publicdeploymentapis.NewDataServiceDeploymentServiceClient(deployment.ApiClientV2)

	return ctx, depClient, token, nil
}

func (deployment *PdsGrpc) GetDeployment(deploymentId string) (*PDSDeploymentResponse, error) {
	depResponse := PDSDeploymentResponse{
		Get: V1Deployment{},
	}
	ctx, client, _, err := deployment.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	getRequest := &publicdeploymentapis.GetDataServiceDeploymentRequest{
		Id: deploymentId,
	}
	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)
	apiResponse, err := client.GetDataServiceDeployment(ctx, getRequest)
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while getting the deployment: %v\n", err)
	}
	err = copier.Copy(&depResponse, apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while copying the response:%v\n", err)
	}
	return &depResponse, nil
}

func (deployment *PdsGrpc) DeleteDeployment(deploymentId string) error {
	//depResponse := WorkFlowResponse{}
	ctx, client, _, err := deployment.getDeploymentClient()
	if err != nil {
		return fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	deleteRequest := &publicdeploymentapis.DeleteDataServiceDeploymentRequest{
		Id: deploymentId,
	}

	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)
	apiResponse, err := client.DeleteDataServiceDeployment(ctx, deleteRequest)
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return fmt.Errorf("Error while deleting the deployment: %v\n", err)
	}
	//err = copier.Copy(&depResponse, apiResponse)
	//if err != nil {
	//	return nil, fmt.Errorf("Error while copying the response:%v\n", err)
	//}
	return nil
}

func (deployment *PdsGrpc) ListDeployment(projectId string) (*PDSDeploymentResponse, error) {
	depResponse := PDSDeploymentResponse{
		Get: V1Deployment{},
	}
	ctx, client, _, err := deployment.getDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	listRequest := &publicdeploymentapis.ListDataServiceDeploymentsRequest{
		Pagination: NewPaginationRequest(1, 50),
		Sort:       nil,
	}

	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)
	apiResponse, err := client.ListDataServiceDeployments(ctx, listRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while creating the deployment: %v\n", err)
	}
	err = copier.Copy(&depResponse, apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while copying the response:%v\n", err)
	}

	return &depResponse, nil
}

func (deployment *PdsGrpc) CreateDeployment(createDeploymentRequest *PDSDeploymentRequest) (*PDSDeploymentResponse, error) {
	depResponse := PDSDeploymentResponse{
		Create: V1Deployment{},
	}
	dep := createDeploymentRequest.Create.V1Deployment

	createRequest := &publicdeploymentapis.CreateDataServiceDeploymentRequest{
		NamespaceId: createDeploymentRequest.Create.NamespaceID,
		DataServiceDeployment: &publicdeploymentapis.DataServiceDeployment{
			Meta: nil,
			Config: &publicdeploymentapis.Config{
				References: nil,
				DataServiceDeploymentTopologies: []*deploymenttopology.DataServiceDeploymentTopology{
					{
						Name:        *dep.Meta.Name,
						Description: "",
						Instances:   3,
						ResourceSettings: &deploymenttopology.Template{
							Id:              *dep.Config.DataServiceDeploymentTopologies[0].ResourceSettings.Id,
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
	log.Debugf("Namespace ID: [%s]", createDeploymentRequest.Create.NamespaceID)
	log.Debugf("workflowrequest ResourceTemplateId: [%s]", *createDeploymentRequest.Create.V1Deployment.Config.DataServiceDeploymentTopologies[0].ResourceSettings.Id)

	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)

	apiResponse, err := client.CreateDataServiceDeployment(ctx, createRequest, grpc.PerRPCCredentials(credentials))
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

func (deployment *PdsGrpc) GetDeploymentCredentials(deploymentId string) (string, error) {
	ctx, client, _, err := deployment.getDeploymentClient()
	if err != nil {
		return "", fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	deploymentCredentialsRequest := &publicdeploymentapis.GetDataServiceDeploymentCredentialsRequest{
		Id: deploymentId,
	}

	ctx = WithAccountIDMetaCtx(ctx, deployment.AccountId)
	apiResponse, err := client.GetDataServiceDeploymentCredentials(ctx, deploymentCredentialsRequest)
	if err != nil {
		return "", fmt.Errorf("Error while getting the deployment: %v\n", err)
	}
	return apiResponse.Secret, nil
}
