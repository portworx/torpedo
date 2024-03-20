package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowDataService struct {
	Namespace               WorkflowNamespace
	NamespaceName           string
	DataServiceDeploymentId string
}

func (wfDataService *WorkflowDataService) DeployDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	log.Infof("targetClusterId [%s]", targetClusterId)
	deployment, err := dslibs.DeployDataService(ds, namespace, projectId, targetClusterId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeploymentId = *deployment.PDSDeployment.V1Deployment.Meta.Uid
	return deployment, nil
}

func (wfDataService *WorkflowDataService) UpdateDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	deployment, err := dslibs.UpdateDataService(ds, wfDataService.Namespace.Namespaces[wfDataService.NamespaceName], wfDataService.Namespace.TargetCluster.Project.ProjectId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeploymentId = *deployment.PDSDeployment.V1Deployment.Meta.Uid
	return deployment, nil
}
