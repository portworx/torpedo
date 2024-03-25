package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowDataService struct {
	Namespace             WorkflowNamespace
	PDSTemplates          CustomTemplates
	NamespaceName         string
	DataServiceDeployment map[string]string
}

func (wfDataService *WorkflowDataService) DeployDataService(ds dslibs.PDSDataService, validateDeployment bool) (*automationModels.WorkFlowResponse, error) {
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplate[serviceConfigTempID]
	resConfigId := wfDataService.PDSTemplates.ResourceTemplate[resourceTempID]
	stConfigId := wfDataService.PDSTemplates.StorageTemplate[storageTempID]
	log.Infof("targetClusterId [%s]", targetClusterId)
	deployment, err := dslibs.DeployDataService(ds, namespace, projectId, targetClusterId, appConfigId, resConfigId, stConfigId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeployment[*deployment.PDSDeployment.V1Deployment.Meta.Name] = *deployment.PDSDeployment.V1Deployment.Meta.Uid

	if validateDeployment {
		err = dslibs.ValidateDataServiceDeployment(wfDataService.DataServiceDeployment, namespace)
		if err != nil {
			return nil, err
		}
	}

	return deployment, nil
}

func (wfDataService *WorkflowDataService) UpdateDataService(ds dslibs.PDSDataService, validateDeployment bool) (*automationModels.WorkFlowResponse, error) {
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplate[serviceConfigTempID]
	resConfigId := wfDataService.PDSTemplates.ResourceTemplate[resourceTempID]
	stConfigId := wfDataService.PDSTemplates.StorageTemplate[storageTempID]
	log.Infof("targetClusterId [%s]", targetClusterId)

	deployment, err := dslibs.UpdateDataService(ds, namespace, projectId, appConfigId, resConfigId, stConfigId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeployment[*deployment.PDSDeployment.V1Deployment.Meta.Name] = *deployment.PDSDeployment.V1Deployment.Meta.Uid
	if validateDeployment {
		err = dslibs.ValidateDataServiceDeployment(wfDataService.DataServiceDeployment, namespace)
		if err != nil {
			return nil, err
		}
	}
	return deployment, nil
}
