package platformLibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

func ListAvailableApplicationsForTenant(clusterId string, tenantId string) ([]automationModels.WorkFlowResponse, error) {

	pdsAppRequest := automationModels.WorkFlowRequest{
		PDSApplication: automationModels.PDSApplicaition{
			ListAvailableAppsForTenant: automationModels.PlatformListAvailableAppsForTenant{},
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

func InstallApplication(applicationName string, clusterId string) (*automationModels.WorkFlowResponse, error) {
	pdsAppRequest := automationModels.WorkFlowRequest{
		PDSApplication: automationModels.PDSApplicaition{
			Install: automationModels.PDSApplicationInstall{},
		},
	}
	pdsAppRequest.PDSApplication.Install.ClusterId = clusterId
	pdsAppRequest.PDSApplication.Install.V1Application1 = &automationModels.V1Application1{
		Meta: &automationModels.V1Meta{
			Name: &applicationName,
		},
	}

	installApp, err := v2Components.Platform.InstallApplication(&pdsAppRequest)
	if err != nil {
		return installApp, err
	}

	return installApp, nil
}
