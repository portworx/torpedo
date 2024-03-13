package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

func ListAvailableApplicationsForTenant(clusterId string, tenantId string) ([]apiStructs.WorkFlowResponse, error) {

	pdsAppRequest := apiStructs.WorkFlowRequest{
		PDSApplication: apiStructs.PDSApplicaition{
			ListAvailableAppsForTenant: apiStructs.PlatformListAvailableAppsForTenant{},
		},
	}
	pdsAppRequest.PDSApplication.ListAvailableAppsForTenant.ClusterId = clusterId
	pdsAppRequest.PDSApplication.ListAvailableAppsForTenant.TenantId = tenantId
	if err != nil {
		return nil, fmt.Errorf("Failed to get Context: %v\n", err)
	}
	availableApps, err := v2Components.Platform.ListAvailableApplicationsForTenant(&pdsAppRequest)

	if err != nil {
		return nil, err
	}

	return availableApps, nil
}

func InstallApplication(applicationName string, applicationVersion string, clusterId string) (*apiStructs.WorkFlowResponse, error) {
	pdsAppRequest := apiStructs.WorkFlowRequest{
		PDSApplication: apiStructs.PDSApplicaition{
			Install: apiStructs.PDSApplicationInstall{},
		},
	}

	pdsAppRequest.PDSApplication.Install.ClusterId = clusterId
	pdsAppRequest.PDSApplication.Install.V1Application1 = &apiStructs.V1Application1{
		Meta: &apiStructs.V1Meta{
			Name: &applicationName,
		},
		Config: &apiStructs.AppConfig{
			Version: applicationVersion,
		},
	}

	installApp, err := v2Components.Platform.InstallApplication(&pdsAppRequest)
	if err != nil {
		return installApp, err
	}

	return installApp, nil
}
