package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
)

type WorkflowDataService struct {
	Namespace               WorkflowNamespace
	NamespaceName           string
	DataServiceDeploymentId string
}

func (wfDataService *WorkflowDataService) DeployDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	deployment, err := dslibs.DeployDataService(ds, wfDataService.Namespace.Namespaces[wfDataService.NamespaceName], wfDataService.Namespace.TargetCluster.Project.ProjectId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeploymentId = *deployment.PDSDeployment.V1Deployment.Meta.Uid
	return deployment, nil
}

func (wfDataService *WorkflowDataService) UpdateDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	deployment, err := dslibs.UpdateDataService(ds)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}
