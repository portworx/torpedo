package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
)

func DeployDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	deployment, err := dslibs.DeployDataService(ds)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func UpdateDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	deployment, err := dslibs.UpdateDataService(ds)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}
