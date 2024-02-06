package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

type ApplicationsV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (app *ApplicationsV2) GetClient() (context.Context, *platformV2.ApplicationServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	app.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	app.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = app.AccountID
	client := app.ApiClientV2.ApplicationServiceAPI

	return ctx, client, nil
}

// ListAllApplicationsInCluster lists all application based on cluster id
func (app *ApplicationsV2) ListAllApplicationsInCluster(clusterId string) ([]platformV2.V1Application1, error) {
	ctx, appClient, err := app.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceListApplications(ctx, clusterId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceListApplications`: %v\n.Full HTTP response: %v", err, res)
	}
	return appModels.Applications, nil
}

// ListAvailableApplicationsForTenant lists all application available across tenant
func (app *ApplicationsV2) ListAvailableApplicationsForTenant(tenantId string) ([]platformV2.V1Application, error) {
	ctx, appClient, err := app.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceListAvailableApplications(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceListAvailableApplications`: %v\n.Full HTTP response: %v", err, res)
	}
	return appModels.Applications, nil
}

// GetApplicationAtClusterLevel gets the app model by its appid and the clusterId its installed in
func (app *ApplicationsV2) GetApplicationAtClusterLevel(appId string, clusterId string) (*platformV2.V1Application1, error) {
	ctx, appClient, err := app.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceGetApplication2(ctx, clusterId, appId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceGetApplication2`: %v\n.Full HTTP response: %v", err, res)
	}
	return appModels, nil
}

// GetApplicationByAppId gets the app model by its appid
func (app *ApplicationsV2) GetApplicationByAppId(appId string) (*platformV2.V1Application1, error) {
	ctx, appClient, err := app.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceGetApplication(ctx, appId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceGetApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	return appModels, nil
}

// InstallApplication installs the app model on given clusterId
func (app *ApplicationsV2) InstallApplication(clusterId string) (*platformV2.V1Application1, error) {
	ctx, appClient, err := app.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	appModels, res, err := appClient.ApplicationServiceInstallApplication(ctx, clusterId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceInstallApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	return appModels, nil
}

// UninstallApplicationByAppId uninstalls the app model by given appId
func (app *ApplicationsV2) UninstallApplicationByAppId(appId string) (*status.Response, error) {
	ctx, appClient, err := app.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := appClient.ApplicationServiceUninstallApplication(ctx, appId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceUninstallApplication`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}

// UninstallAppByAppIdClusterId uninstalls the app model by given appId and clusterId
func (app *ApplicationsV2) UninstallAppByAppIdClusterId(appId string, clusterId string) (*status.Response, error) {
	ctx, appClient, err := app.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := appClient.ApplicationServiceUninstallApplication2(ctx, clusterId, appId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApplicationServiceUninstallApplication2`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
