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

// getApplicationClientForTenant updates the header with bearer token and returns the new client
func (ApplicationGrpcV1 *PlatformGrpc) getAppClientForTenant() (context.Context, publicAppsApitenant.ApplicationServiceClient, string, error) {
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
func (ApplicationGrpcV1 *PlatformGrpc) getAppClientForCluster() (context.Context, publicAppsApicluster.ApplicationServiceClient, string, error) {
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
func (ApplicationGrpcV1 *PlatformGrpc) ListAvailableApplicationsForTenant(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForTenant()
	applicationResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *publicAppsApitenant.ListAvailableApplicationsRequest
	firstPageRequest.TenantId = listReq.PDSApplication.ListAvailableAppsForTenant.TenantId
	if err != nil {
		return nil, err
	}
	appModels, err := appClient.ListAvailableApplications(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting while calling `ListAvailableApplicationsRequest`: %v\n", err)
	}
	log.Infof("Value of applications - [%v]", appModels)
	err = copier.Copy(&applicationResponse, appModels.Applications)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", applicationResponse)
	return applicationResponse, nil
}

// ListAllApplicationsInCluster lists all application based on clusterId id
func (ApplicationGrpcV1 *PlatformGrpc) ListAllApplicationsInCluster(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	applicationResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest2 *publicAppsApicluster.ListApplicationsRequest
	err = copier.Copy(&firstPageRequest2, listReq)
	if err != nil {
		return nil, err
	}
	appModels, err := appClient.ListApplications(ctx, firstPageRequest2, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting while calling `ListAvailableApplicationsRequest`: %v\n", err)
	}
	log.Infof("Value of applications - [%v]", appModels)
	err = copier.Copy(&applicationResponse, appModels.Applications)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", applicationResponse)
	return applicationResponse, nil
}

// GetApplicationAtClusterLevel gets the app model by its appid and the clusterId its installed in
func (ApplicationGrpcV1 *PlatformGrpc) GetApplicationAtClusterLevel(getReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var getAppRequest *publicAppsApicluster.GetApplicationRequest
	err = copier.Copy(&getAppRequest, getReq)
	if err != nil {
		return nil, err
	}
	appModel, err := appClient.GetApplication(ctx, getAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetApplication`: %v\n", err)
	}
	log.Infof("Value of app - [%v]", appModel)
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of app after copy - [%v]", appResponse)
	return &appResponse, nil
}

// GetApplicationByAppId gets the app model by its appid
func (ApplicationGrpcV1 *PlatformGrpc) GetApplicationByAppId(getReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var getAppRequest *publicAppsApicluster.GetApplicationRequest
	err = copier.Copy(&getAppRequest, getReq)
	if err != nil {
		return nil, err
	}
	appModel, err := appClient.GetApplication(ctx, getAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetApplication`: %v\n", err)
	}
	log.Infof("Value of app - [%v]", appModel)
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of app after copy - [%v]", appResponse)
	return &appResponse, nil
}

// InstallApplication installs the app model on given clusterId
func (ApplicationGrpcV1 *PlatformGrpc) InstallApplication(installRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var installAppReq *publicAppsApicluster.InstallApplicationRequest
	err = copier.Copy(&installAppReq, installRequest)
	if err != nil {
		return nil, err
	}
	appModel, err := appClient.InstallApplication(ctx, installAppReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `InstallApplication`: %v\n", err)
	}
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// UninstallApplicationByAppId uninstalls the app model by given appId
func (ApplicationGrpcV1 *PlatformGrpc) UninstallApplicationByAppId(uninstallReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var uninstallAppReq *publicAppsApicluster.UninstallApplicationRequest
	err = copier.Copy(&uninstallAppReq, uninstallReq)
	if err != nil {
		return nil, err
	}
	appModel, err := appClient.UninstallApplication(ctx, uninstallAppReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `UninstallApplication`: %v\n", err)
	}
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// UninstallAppByAppIdClusterId uninstalls the app model by given appId and clusterId
func (ApplicationGrpcV1 *PlatformGrpc) UninstallAppByAppIdClusterId(uninstallReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, appClient, _, err := ApplicationGrpcV1.getAppClientForCluster()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var uninstallAppReq *publicAppsApicluster.UninstallApplicationRequest
	err = copier.Copy(&uninstallAppReq, uninstallReq)
	if err != nil {
		return nil, err
	}
	appModel, err := appClient.UninstallApplication(ctx, uninstallAppReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling api `UninstallApplication`: %v\n", err)
	}
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}
