package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetAppClient updates the header with bearer token and returns the new client
func (ns *PLATFORM_API_V1) GetAppClient() (context.Context, *platformv1.ApplicationServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ns.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ns.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = ns.AccountID
	client := ns.ApiClientV1.ApplicationServiceAPI

	return ctx, client, nil
}

// ListAllApplicationsInCluster lists all application based on cluster id
func (ns *PLATFORM_API_V1) ListAllApplicationsInCluster(clusterId string) ([]ApiResponse, error) {
	ctx, appClient, err := ns.GetAppClient()
	applicationResponse := []ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceListApplications(ctx, clusterId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceListApplications`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of applications - [%v]", appModels)
	copier.Copy(&applicationResponse, appModels.Applications)
	log.Infof("Value of applications after copy - [%v]", applicationResponse)
	return applicationResponse, nil
}

// ListAvailableApplicationsForTenant lists all application available across tenant
func (ns *PLATFORM_API_V1) ListAvailableApplicationsForTenant(tenantId string) ([]ApiResponse, error) {
	ctx, appClient, err := ns.GetAppClient()
	applicationResponse := []ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceListAvailableApplications(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceListAvailableApplications`: %v\n.Full HTTP response: %v", err, res)
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
func (ns *PLATFORM_API_V1) GetApplicationAtClusterLevel(appId string, clusterId string) (*ApiResponse, error) {
	ctx, appClient, err := ns.GetAppClient()
	appResponse := ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	var getRequest platformv1.ApiApplicationServiceGetApplication2Request
	getRequest = getRequest.ApiService.ApplicationServiceGetApplication2(ctx, clusterId, appId)
	appModel, res, err := appClient.ApplicationServiceGetApplication2Execute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceGetApplication2`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of applications - [%v]", appModel)
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// GetApplicationByAppId gets the app model by its appid
func (ns *PLATFORM_API_V1) GetApplicationByAppId(appId string) (*ApiResponse, error) {
	ctx, appClient, err := ns.GetAppClient()
	appResponse := ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	var getRequest platformv1.ApiApplicationServiceGetApplicationRequest
	getRequest = getRequest.ApiService.ApplicationServiceGetApplication(ctx, appId)
	appModel, res, err := appClient.ApplicationServiceGetApplicationExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceGetApplication2`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of applications - [%v]", appModel)
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// InstallApplication installs the app model on given clusterId
func (ns *PLATFORM_API_V1) InstallApplication(installRequest platformv1.ApiApplicationServiceInstallApplicationRequest, clusterId string) (*ApiResponse, error) {
	ctx, appClient, err := ns.GetAppClient()
	appResponse := ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	installRequest = installRequest.ApiService.ApplicationServiceInstallApplication(ctx, clusterId)
	appModel, res, err := appClient.ApplicationServiceInstallApplicationExecute(installRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceInstallApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&appResponse, appModel)
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// UninstallApplicationByAppId uninstalls the app model by given appId
func (ns *PLATFORM_API_V1) UninstallApplicationByAppId(appId string) error {
	ctx, appClient, err := ns.GetAppClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := appClient.ApplicationServiceUninstallApplication(ctx, appId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `ApplicationServiceUninstallApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}

// UninstallAppByAppIdClusterId uninstalls the app model by given appId and clusterId
func (ns *PLATFORM_API_V1) UninstallAppByAppIdClusterId(appId string, clusterId string) error {
	ctx, appClient, err := ns.GetAppClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := appClient.ApplicationServiceUninstallApplication2(ctx, clusterId, appId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `ApplicationServiceUninstallApplication2`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}
