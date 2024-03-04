package api

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
func (applications *PLATFORM_API_V1) GetAppClient() (context.Context, *platformv1.ApplicationServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	applications.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	applications.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = applications.AccountID
	client := applications.ApiClientV1.ApplicationServiceAPI

	return ctx, client, nil
}

// ListAllApplicationsInCluster lists all application based on cluster id
func (applications *PLATFORM_API_V1) ListAllApplicationsInCluster(appRequest *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, appClient, err := applications.GetAppClient()
	applicationResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceListApplications(ctx, appRequest.ClusterId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceListApplications`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of applications - [%v]", appModels)
	err = copier.Copy(&applicationResponse, appModels.Applications)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", applicationResponse)
	return applicationResponse, nil
}

// ListAvailableApplicationsForTenant lists all application available across tenant
func (applications *PLATFORM_API_V1) ListAvailableApplicationsForTenant(appRequest *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, appClient, err := applications.GetAppClient()
	applicationResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceListAvailableApplications(ctx, appRequest.TenantId).Execute()
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
func (applications *PLATFORM_API_V1) GetApplicationAtClusterLevel(appReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, appClient, err := applications.GetAppClient()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var getRequest platformv1.ApiApplicationServiceGetApplication2Request
	getRequest = getRequest.ApiService.ApplicationServiceGetApplication2(ctx, appReq.ClusterId, appReq.PdsAppId)
	appModel, res, err := appClient.ApplicationServiceGetApplication2Execute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceGetApplication2`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of applications - [%v]", appModel)
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// GetApplicationByAppId gets the app model by its appid
func (applications *PLATFORM_API_V1) GetApplicationByAppId(appReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, appClient, err := applications.GetAppClient()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var getRequest platformv1.ApiApplicationServiceGetApplicationRequest
	getRequest = getRequest.ApiService.ApplicationServiceGetApplication(ctx, appReq.PdsAppId)
	appModel, res, err := appClient.ApplicationServiceGetApplicationExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceGetApplication2`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of applications - [%v]", appModel)
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// InstallApplication installs the app model on given clusterId
func (applications *PLATFORM_API_V1) InstallApplication(appInstallRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	var installRequest platformv1.ApiApplicationServiceInstallApplicationRequest
	_, appClient, err := applications.GetAppClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appResponse := WorkFlowResponse{}
	err = copier.Copy(&installRequest, appInstallRequest)
	if err != nil {
		return nil, err
	}
	appModel, res, err := appClient.ApplicationServiceInstallApplicationExecute(installRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceInstallApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// UninstallApplicationByAppId uninstalls the app model by given appId
func (applications *PLATFORM_API_V1) UninstallApplicationByAppId(appUninstallRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	var uninstallReq platformv1.ApiApplicationServiceUninstallApplicationRequest
	_, appClient, err := applications.GetAppClient()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	err = copier.Copy(&uninstallReq, appUninstallRequest)
	if err != nil {
		return nil, err
	}
	appModel, res, err := appClient.ApplicationServiceUninstallApplicationExecute(uninstallReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceUninstallApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}

// UninstallAppByAppIdClusterId uninstalls the app model by given appId and clusterId
func (applications *PLATFORM_API_V1) UninstallAppByAppIdClusterId(appUninstallRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	var uninstallReq platformv1.ApiApplicationServiceUninstallApplication2Request
	_, appClient, err := applications.GetAppClient()
	appResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	err = copier.Copy(&uninstallReq, appUninstallRequest)
	if err != nil {
		return nil, err
	}
	appModel, res, err := appClient.ApplicationServiceUninstallApplication2Execute(uninstallReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceUninstallApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&appResponse, appModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of applications after copy - [%v]", appResponse)
	return &appResponse, nil
}
