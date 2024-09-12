package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

func ListAvailableApplicationsForCluster(clusterId string) (*automationModels.WorkFlowResponse, error) {
	pdsAppRequest := automationModels.WorkFlowRequest{
		PDSApplication: automationModels.PDSApplication{
			ListAvailableAppsForCluster: automationModels.PlatformListAvailableAppsForCluster{
				ClusterId: clusterId,
			},
		},
	}

	availableApps, err := v2Components.Platform.ListAllApplicationsInCluster(&pdsAppRequest)
	if err != nil {
		return nil, err
	}

	return availableApps, nil
}

func InstallApplication(applicationName string, clusterId string) (*automationModels.WorkFlowResponse, error) {
	pdsAppRequest := automationModels.WorkFlowRequest{
		PDSApplication: automationModels.PDSApplication{
			Install: automationModels.PDSApplicationInstall{},
		},
	}
	pdsAppRequest.PDSApplication.Install.ClusterId = clusterId
	pdsAppRequest.PDSApplication.Install.V1Application1 = &automationModels.V1Application{
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

func UpdateApplication(applicationID string, enableTLS bool) (*automationModels.WorkFlowResponse, error) {
	pdsAppRequest := automationModels.WorkFlowRequest{
		PDSApplication: automationModels.PDSApplication{
			Update: automationModels.PDSApplicationUpdate{
				Application: automationModels.V1Application{
					Meta: &automationModels.V1Meta{
						Uid: &applicationID,
					},
					Config: &automationModels.AppConfig{
						Pds: &automationModels.V1PDSProperties{
							Global: &automationModels.PDSPropertiesGlobal{
								DataServiceTlsEnabled: &enableTLS,
							},
						},
					},
					Status: nil,
				},
			},
		},
	}
	updateApp, err := v2Components.Platform.UpdateApplicationInCluster(&pdsAppRequest)
	if err != nil {
		return updateApp, err
	}
	return updateApp, nil
}
