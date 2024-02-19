package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicAppsApicluster "github.com/pure-px/apis/public/portworx/platform/targetcluster/application/apiv1"
	publicAppsApitenant "github.com/pure-px/apis/public/portworx/platform/tenant/application/apiv1"
	"google.golang.org/grpc"
)

type ApplicationGrpc struct {
	ApiClientV1 *grpc.ClientConn
}

// getApplicationClientForTenant updates the header with bearer token and returns the new client
func (ApplicationGrpcV1 *ApplicationGrpc) getAppClientForTenant() (context.Context, publicAppsApitenant.ApplicationServiceClient, string, error) {
	log.Infof("Creating client from grpc package")

	var ApplicationGrpcV1Client publicAppsApitenant.ApplicationServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	ApplicationGrpcV1Client = publicAppsApitenant.NewApplicationServiceClient(ApplicationGrpcV1.ApiClientV1)

	return ctx, ApplicationGrpcV1Client, token, nil

}

// getApplicationClientForTenant updates the header with bearer token and returns the new client
func (ApplicationGrpcV1 *ApplicationGrpc) getAppClientForCluster() (context.Context, publicAppsApicluster.ApplicationServiceClient, string, error) {
	log.Infof("Creating client from grpc package")

	var ApplicationGrpcV1Client publicAppsApicluster.ApplicationServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	ApplicationGrpcV1Client = publicAppsApicluster.NewApplicationServiceClient(ApplicationGrpcV1.ApiClientV1)

	return ctx, ApplicationGrpcV1Client, token, nil

}

// ListAvailableApplicationsForTenant lists all application based on tenant id
func (ApplicationGrpcV1 *ApplicationGrpc) ListAvailableApplicationsForTenant(tenantId string) ([]WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForTenant()
	applicationResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	firstPageRequest := &publicAppsApitenant.ListAvailableApplicationsRequest{TenantId: tenantId}
	appModels, err := appClient.ListAvailableApplications(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting while calling `ListAvailableApplicationsRequest`: %v\n", err)
	}
	log.Infof("Value of applications - [%v]", appModels)
	copier.Copy(&applicationResponse, appModels.Applications)
	log.Infof("Value of applications after copy - [%v]", applicationResponse)
	return applicationResponse, nil
}

// ListAllApplicationsInCluster lists all application based on clusterId id
func (ApplicationGrpcV1 *ApplicationGrpc) ListAllApplicationsInCluster(clusterId string) ([]WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	applicationResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	firstPageRequest2 := &publicAppsApicluster.ListApplicationsRequest{
		ClusterId:  clusterId,
		Pagination: NewPaginationRequest(1, 50),
	}
	appModels, err := appClient.ListApplications(ctx, firstPageRequest2, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting while calling `ListAvailableApplicationsRequest`: %v\n", err)
	}
	log.Infof("Value of applications - [%v]", appModels)
	copier.Copy(&applicationResponse, appModels.Applications)
	log.Infof("Value of applications after copy - [%v]", applicationResponse)
	return applicationResponse, nil
}

// GetApplicationAtClusterLevel gets the app model by its appid and the clusterId its installed in
func (ApplicationGrpcV1 *ApplicationGrpc) GetApplicationAtClusterLevel(clusterId, appId string) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	getRequest := &publicAppsApicluster.GetApplicationRequest{
		ClusterId: clusterId,
		Id:        appId,
	}
	appModel, err := appClient.GetApplication(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetApplication`: %v\n", err)
	}
	log.Infof("Value of app - [%v]", appModel)
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of app after copy - [%v]", appResponse)
	return &appResponse, nil

}

// GetApplicationByAppId gets the app model by its appid
func (ApplicationGrpcV1 *ApplicationGrpc) GetApplicationByAppId(appId string) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	getRequest := &publicAppsApicluster.GetApplicationRequest{Id: appId}
	appModel, err := appClient.GetApplication(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetApplication`: %v\n", err)
	}
	log.Infof("Value of app - [%v]", appModel)
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of app after copy - [%v]", appResponse)
	return &appResponse, nil
}

// InstallApplication installs the app model on given clusterId
func (ApplicationGrpcV1 *ApplicationGrpc) InstallApplication(installRequest *publicAppsApicluster.InstallApplicationRequest, clusterId string) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	installRequest = &publicAppsApicluster.InstallApplicationRequest{ClusterId: clusterId}
	appModel, err := appClient.InstallApplication(ctx, installRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `InstallApplication`: %v\n", err)
	}
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// UninstallApplicationByAppId uninstalls the app model by given appId
func (ApplicationGrpcV1 *ApplicationGrpc) UninstallApplicationByAppId(appId string, uninstallReq *publicAppsApicluster.UninstallApplicationRequest) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	uninstallReq = &publicAppsApicluster.UninstallApplicationRequest{Id: appId}
	appModel, err := appClient.UninstallApplication(ctx, uninstallReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `UninstallApplication`: %v\n", err)
	}
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// UninstallAppByAppIdClusterId uninstalls the app model by given appId and clusterId
func (ApplicationGrpcV1 *ApplicationGrpc) UninstallAppByAppIdClusterId(appId string, clusterId string, uninstallReq *publicAppsApicluster.UninstallApplicationRequest) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	uninstallReq = &publicAppsApicluster.UninstallApplicationRequest{
		ClusterId: clusterId,
		Id:        appId,
	}
	appModel, err := appClient.UninstallApplication(ctx, uninstallReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `UninstallApplication`: %v\n", err)
	}
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}
